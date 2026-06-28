package provider

import (
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/ccapi/internal/payment"
)

func TestBuildTianqueSignStringPreservesReqDataOrder(t *testing.T) {
	t.Parallel()

	reqData := newTianqueOrderedMap()
	reqData.Set("mno", "mno-1")
	reqData.Set("ordNo", "order-1")
	reqData.Set("amt", "12.30")

	params := newTianqueOrderedMap()
	params.Set("version", "1.2")
	params.Set("orgId", "org-1")
	params.Set("reqData", reqData)
	params.Set("sign", "ignored")

	got := buildTianqueSignString(params)
	want := `orgId=org-1&reqData={"mno":"mno-1","ordNo":"order-1","amt":"12.30"}&version=1.2`
	if got != want {
		t.Fatalf("buildTianqueSignString() = %q, want %q", got, want)
	}
}

func TestTianqueCreatePaymentSendsSignedActiveScanRequest(t *testing.T) {
	t.Parallel()

	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/activeScan" {
			t.Fatalf("path = %q, want /order/activeScan", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &gotPayload); err != nil {
			t.Fatalf("unmarshal request: %v", err)
		}
		_, _ = w.Write([]byte(`{"code":"0000","respData":{"bizCode":"0000","payUrl":"https://pay.example/q","uuid":"uuid-1"}}`))
	}))
	defer server.Close()

	prov := newTestTianque(t, server.URL)
	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "order-1",
		Amount:      "12.30",
		PaymentType: payment.TypeAlipay,
		Subject:     "CCAPI recharge",
		NotifyURL:   "https://merchant.example.com/api/v1/payment/webhook/tianque",
		ClientIP:    "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if resp.TradeNo != "uuid-1" || resp.PayURL != "https://pay.example/q" || resp.QRCode != "https://pay.example/q" {
		t.Fatalf("response = %+v", resp)
	}
	if gotPayload["sign"] == "" {
		t.Fatalf("missing sign in payload: %#v", gotPayload)
	}
	reqData, ok := gotPayload["reqData"].(map[string]any)
	if !ok {
		t.Fatalf("reqData type = %T", gotPayload["reqData"])
	}
	for key, want := range map[string]string{
		"mno":       "mno-1",
		"ordNo":     "order-1",
		"amt":       "12.30",
		"payType":   "ALIPAY",
		"subject":   "CCAPI recharge",
		"notifyUrl": "https://merchant.example.com/api/v1/payment/webhook/tianque",
		"trmIp":     "203.0.113.10",
	} {
		if got, _ := reqData[key].(string); got != want {
			t.Fatalf("reqData[%s] = %q, want %q (reqData=%v)", key, got, want, reqData)
		}
	}
}

func TestTianqueVerifyNotification(t *testing.T) {
	t.Parallel()

	prov := newTestTianque(t, "https://openapi-test.tianquetech.com")
	notification, err := prov.VerifyNotification(context.Background(), `{"code":"0000","respData":{"bizCode":"0000","ordNo":"order-1","tranSts":"SUCCESS","amt":"12.30","uuid":"uuid-1","mno":"mno-1"}}`, nil)
	if err != nil {
		t.Fatalf("VerifyNotification returned error: %v", err)
	}
	if notification.OrderID != "order-1" || notification.TradeNo != "uuid-1" || notification.Status != payment.NotificationStatusSuccess {
		t.Fatalf("notification = %+v", notification)
	}
	if notification.Metadata["mno"] != "mno-1" {
		t.Fatalf("metadata = %#v", notification.Metadata)
	}
}

func newTestTianque(t *testing.T, apiBase string) *Tianque {
	t.Helper()
	prov, err := NewTianque("test-instance", map[string]string{
		"orgId":      "org-1",
		"mno":        "mno-1",
		"privateKey": testTianquePrivateKeyPEM(t),
		"apiBase":    apiBase,
		"version":    "1.2",
	})
	if err != nil {
		t.Fatalf("NewTianque: %v", err)
	}
	return prov
}

func testTianquePrivateKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(crand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal RSA key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}
