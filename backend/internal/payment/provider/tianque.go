package provider

import (
	"bytes"
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/ccapi/internal/payment"
)

const (
	tianqueDefaultAPIBase      = "https://openapi-test.tianquetech.com"
	tianqueDefaultVersion      = "1.2"
	tianqueHTTPTimeout         = 30 * time.Second
	tianqueMaxResponseSize     = 1 << 20
	tianqueMaxErrorSummary     = 512
	tianquePayTypeAlipay       = "ALIPAY"
	tianquePayTypeWechat       = "WECHAT"
	tianqueTradeSourceActiveQR = "01"
	tianqueCodeSuccess         = "0000"
	tianqueBizCodeRefundAccept = "2002"
	tianqueTranStatusSuccess   = "SUCCESS"
)

// Tianque implements payment.Provider for 随行付 / 天阙聚合支付.
type Tianque struct {
	instanceID string
	config     map[string]string
	privateKey *rsa.PrivateKey
	httpClient *http.Client
}

// NewTianque creates a new 随行付 provider.
// config keys: orgId, mno, privateKey, apiBase, version, notifyUrl
func NewTianque(instanceID string, config map[string]string) (*Tianque, error) {
	for _, k := range []string{"orgId", "mno", "privateKey", "apiBase"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("tianque config missing required key: %s", k)
		}
	}
	privateKey, err := loadTianquePrivateKey(config["privateKey"])
	if err != nil {
		return nil, fmt.Errorf("load tianque private key: %w", err)
	}
	cfg := cloneStringMap(config)
	cfg["apiBase"] = normalizeTianqueAPIBase(cfg["apiBase"])
	if strings.TrimSpace(cfg["version"]) == "" {
		cfg["version"] = tianqueDefaultVersion
	}
	return &Tianque{
		instanceID: instanceID,
		config:     cfg,
		privateKey: privateKey,
		httpClient: &http.Client{Timeout: tianqueHTTPTimeout},
	}, nil
}

func normalizeTianqueAPIBase(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return tianqueDefaultAPIBase
	}
	if parsed, err := url.Parse(base); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		parsed.RawPath = ""
		parsed.Path = strings.TrimRight(parsed.Path, "/")
		return strings.TrimRight(parsed.String(), "/")
	}
	return strings.TrimRight(base, "/")
}

func (t *Tianque) Name() string        { return "随行付" }
func (t *Tianque) ProviderKey() string { return payment.TypeTianque }
func (t *Tianque) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay, payment.TypeWxpay}
}

func (t *Tianque) MerchantIdentityMetadata() map[string]string {
	if t == nil {
		return nil
	}
	metadata := map[string]string{}
	if orgID := strings.TrimSpace(t.config["orgId"]); orgID != "" {
		metadata["org_id"] = orgID
	}
	if mno := strings.TrimSpace(t.config["mno"]); mno != "" {
		metadata["mno"] = mno
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
}

func (t *Tianque) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	reqData := newTianqueOrderedMap()
	reqData.Set("mno", t.config["mno"])
	reqData.Set("ordNo", req.OrderID)
	reqData.Set("amt", req.Amount)
	reqData.Set("payType", tianquePayType(req.PaymentType))
	reqData.Set("subject", strings.TrimSpace(req.Subject))
	reqData.Set("tradeSource", tianqueTradeSourceActiveQR)
	reqData.Set("trmIp", tianqueClientIP(req.ClientIP))
	if notifyURL := strings.TrimSpace(req.NotifyURL); notifyURL != "" {
		reqData.Set("notifyUrl", notifyURL)
	} else if notifyURL := strings.TrimSpace(t.config["notifyUrl"]); notifyURL != "" {
		reqData.Set("notifyUrl", notifyURL)
	}

	resp, err := t.doRequest(ctx, "/order/activeScan", reqData)
	if err != nil {
		return nil, fmt.Errorf("tianque create payment: %w", err)
	}

	respData, err := tianqueRespData(resp)
	if err != nil {
		return nil, err
	}
	if err := requireTianqueBizSuccess(respData, "随行付下单失败", tianqueCodeSuccess); err != nil {
		return nil, err
	}

	payURL, _ := respData["payUrl"].(string)
	uuid, _ := respData["uuid"].(string)
	return &payment.CreatePaymentResponse{
		TradeNo: uuid,
		PayURL:  payURL,
		QRCode:  payURL,
	}, nil
}

func tianquePayType(paymentType string) string {
	if payment.GetBasePaymentType(paymentType) == payment.TypeAlipay {
		return tianquePayTypeAlipay
	}
	return tianquePayTypeWechat
}

func tianqueClientIP(clientIP string) string {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return "127.0.0.1"
	}
	return clientIP
}

func (t *Tianque) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	orderID := strings.TrimSpace(tradeNo)
	if orderID == "" {
		return nil, fmt.Errorf("tianque query order: missing order id")
	}
	reqData := newTianqueOrderedMap()
	reqData.Set("mno", t.config["mno"])
	reqData.Set("ordNo", orderID)

	resp, err := t.doRequest(ctx, "/query/tradeQuery", reqData)
	if err != nil {
		return nil, fmt.Errorf("tianque query order: %w", err)
	}
	respData, err := tianqueRespData(resp)
	if err != nil {
		return nil, err
	}
	if err := requireTianqueBizSuccess(respData, "随行付查询失败", tianqueCodeSuccess); err != nil {
		return nil, err
	}

	status := payment.ProviderStatusPending
	if strings.EqualFold(strings.TrimSpace(tianqueString(respData, "tranSts")), tianqueTranStatusSuccess) {
		status = payment.ProviderStatusPaid
	}
	amount, _ := parseAmountFloat(tianqueString(respData, "amt"))
	return &payment.QueryOrderResponse{
		TradeNo:  tianqueFirstNonEmpty(tianqueString(respData, "uuid"), orderID),
		Status:   status,
		Amount:   amount,
		Metadata: t.tianqueMetadata(respData),
	}, nil
}

func (t *Tianque) VerifyNotification(_ context.Context, rawBody string, _ map[string]string) (*payment.PaymentNotification, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(rawBody), &payload); err != nil {
		return nil, fmt.Errorf("tianque parse notify: %w", err)
	}
	if code := tianqueString(payload, "code"); code != "" && code != tianqueCodeSuccess {
		return nil, fmt.Errorf("tianque notify code=%s msg=%s", code, tianqueString(payload, "msg"))
	}

	respData, err := tianqueRespData(payload)
	if err != nil {
		return nil, err
	}
	orderID := strings.TrimSpace(tianqueString(respData, "ordNo"))
	if orderID == "" {
		return nil, fmt.Errorf("tianque notify missing ordNo")
	}

	status := payment.ProviderStatusFailed
	if tianqueString(respData, "bizCode") == tianqueCodeSuccess &&
		strings.EqualFold(strings.TrimSpace(tianqueString(respData, "tranSts")), tianqueTranStatusSuccess) {
		status = payment.NotificationStatusSuccess
	}
	amount, _ := parseAmountFloat(tianqueString(respData, "amt"))
	return &payment.PaymentNotification{
		TradeNo:  tianqueFirstNonEmpty(tianqueString(respData, "uuid"), orderID),
		OrderID:  orderID,
		Amount:   amount,
		Status:   status,
		RawData:  rawBody,
		Metadata: t.tianqueMetadata(respData),
	}, nil
}

func (t *Tianque) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	orderID := strings.TrimSpace(req.OrderID)
	if orderID == "" {
		return nil, fmt.Errorf("tianque refund missing order id")
	}
	refundID := fmt.Sprintf("%s-refund-%d", orderID, time.Now().UnixNano())
	reqData := newTianqueOrderedMap()
	reqData.Set("mno", t.config["mno"])
	reqData.Set("ordNo", refundID)
	reqData.Set("origOrderNo", orderID)
	reqData.Set("amt", req.Amount)

	resp, err := t.doRequest(ctx, "/order/refund", reqData)
	if err != nil {
		return nil, fmt.Errorf("tianque refund: %w", err)
	}
	respData, err := tianqueRespData(resp)
	if err != nil {
		return nil, err
	}
	if err := requireTianqueBizSuccess(respData, "随行付退款失败", tianqueCodeSuccess, tianqueBizCodeRefundAccept); err != nil {
		return nil, err
	}
	return &payment.RefundResponse{RefundID: refundID, Status: payment.ProviderStatusSuccess}, nil
}

func (t *Tianque) doRequest(ctx context.Context, path string, reqData *tianqueOrderedMap) (map[string]any, error) {
	params := newTianqueOrderedMap()
	params.Set("signType", "RSA")
	params.Set("version", t.config["version"])
	params.Set("orgId", t.config["orgId"])
	params.Set("reqId", generateTianqueReqID())
	params.Set("timestamp", time.Now().Format("20060102150405"))
	params.Set("reqData", reqData)

	signature, err := tianqueSign(buildTianqueSignString(params), t.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}
	params.Set("sign", signature)

	body, err := json.Marshal(params.ToMap())
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.config["apiBase"]+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := t.httpClient
	if client == nil {
		client = &http.Client{Timeout: tianqueHTTPTimeout}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, tianqueMaxResponseSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, summarizeTianqueResponse(respBody))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if code := tianqueString(result, "code"); code != "" && code != tianqueCodeSuccess {
		return nil, fmt.Errorf("随行付接口错误: code=%s msg=%s", code, tianqueString(result, "msg"))
	}
	return result, nil
}

func (t *Tianque) tianqueMetadata(respData map[string]any) map[string]string {
	metadata := t.MerchantIdentityMetadata()
	if metadata == nil {
		metadata = map[string]string{}
	}
	for _, key := range []string{"mno", "orgId", "tranSts", "bizCode"} {
		if value := tianqueString(respData, key); value != "" {
			metadata[tianqueMetadataKey(key)] = value
		}
	}
	return metadata
}

func tianqueMetadataKey(key string) string {
	switch key {
	case "orgId":
		return "org_id"
	case "tranSts":
		return "tran_status"
	default:
		return strings.ToLower(key)
	}
}

func tianqueRespData(payload map[string]any) (map[string]any, error) {
	respData, ok := payload["respData"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid tianque response: missing respData")
	}
	return respData, nil
}

func requireTianqueBizSuccess(respData map[string]any, prefix string, acceptedCodes ...string) error {
	bizCode := tianqueString(respData, "bizCode")
	for _, accepted := range acceptedCodes {
		if bizCode == accepted {
			return nil
		}
	}
	return fmt.Errorf("%s: bizCode=%s bizMsg=%s", prefix, bizCode, tianqueString(respData, "bizMsg"))
}

func tianqueString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	switch v := values[key].(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func tianqueFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseAmountFloat(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseFloat(raw, 64)
}

func loadTianquePrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	keyStr = strings.TrimSpace(keyStr)
	if !strings.Contains(keyStr, "-----BEGIN") {
		keyStr = "-----BEGIN PRIVATE KEY-----\n" + keyStr + "\n-----END PRIVATE KEY-----"
	}

	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, fmt.Errorf("PEM decode failed")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		rsaKey, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key failed (PKCS8: %v, PKCS1: %v)", err, err2)
		}
		return rsaKey, nil
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func buildTianqueSignString(params *tianqueOrderedMap) string {
	keys := make([]string, 0)
	for _, k := range params.Keys() {
		if k == "sign" {
			continue
		}
		v := params.Get(k)
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := params.Get(k)
		var str string
		if om, ok := v.(*tianqueOrderedMap); ok {
			str = om.ToJSON()
		} else {
			str = fmt.Sprintf("%v", v)
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, str))
	}
	return strings.Join(parts, "&")
}

func tianqueSign(signString string, privateKey *rsa.PrivateKey) (string, error) {
	hashed := sha1.Sum([]byte(signString))
	sig, err := rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA1, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func generateTianqueReqID() string {
	id := fmt.Sprintf("%x%x", time.Now().UnixNano(), rand.Int64())
	if len(id) > 32 {
		return id[:32]
	}
	return id + strings.Repeat("0", 32-len(id))
}

func summarizeTianqueResponse(body []byte) string {
	summary := strings.Join(strings.Fields(string(body)), " ")
	if summary == "" {
		return "<empty>"
	}
	if len(summary) > tianqueMaxErrorSummary {
		return summary[:tianqueMaxErrorSummary] + "..."
	}
	return summary
}

type tianqueOrderedMap struct {
	keys   []string
	values map[string]any
}

func newTianqueOrderedMap() *tianqueOrderedMap {
	return &tianqueOrderedMap{keys: make([]string, 0), values: make(map[string]any)}
}

func (m *tianqueOrderedMap) Set(key string, value any) {
	if _, exists := m.values[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

func (m *tianqueOrderedMap) Get(key string) any {
	return m.values[key]
}

func (m *tianqueOrderedMap) Keys() []string {
	return m.keys
}

func (m *tianqueOrderedMap) ToJSON() string {
	parts := make([]string, 0, len(m.keys))
	for _, k := range m.keys {
		keyJSON, _ := json.Marshal(k)
		valJSON := ""
		switch val := m.values[k].(type) {
		case *tianqueOrderedMap:
			valJSON = val.ToJSON()
		case string:
			valJSON = `"` + escapeTianqueJSONString(val) + `"`
		default:
			b, _ := json.Marshal(val)
			valJSON = string(b)
		}
		parts = append(parts, string(keyJSON)+":"+valJSON)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (m *tianqueOrderedMap) ToMap() map[string]any {
	result := make(map[string]any, len(m.keys))
	for _, k := range m.keys {
		v := m.values[k]
		if om, ok := v.(*tianqueOrderedMap); ok {
			result[k] = json.RawMessage(om.ToJSON())
		} else {
			result[k] = v
		}
	}
	return result
}

func escapeTianqueJSONString(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

var (
	_ payment.Provider                 = (*Tianque)(nil)
	_ payment.MerchantIdentityProvider = (*Tianque)(nil)
)
