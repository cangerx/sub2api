package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/ccapi/internal/config"
	"github.com/tidwall/gjson"
)

type VideoUpstreamClient interface {
	Create(ctx context.Context, account *Account, template *VideoCallTemplate, body map[string]any) (*VideoUpstreamCreateResult, error)
	Query(ctx context.Context, account *Account, template *VideoCallTemplate, upstreamTaskID string) (*VideoUpstreamQueryResult, error)
	Cancel(ctx context.Context, account *Account, template *VideoCallTemplate, upstreamTaskID string) error
	// RecognizeTemplate asks an OpenAI-style chat model (reached via the given
	// account's base_url + credentials) to turn pasted upstream API docs into a
	// VideoCallTemplate. It is admin-only tooling and fails closed on bad output.
	RecognizeTemplate(ctx context.Context, account *Account, model, doc string) (*VideoCallTemplate, error)
}

type VideoUpstreamCreateResult struct {
	TaskID   string         `json:"task_id"`
	Response map[string]any `json:"response"`
}

type VideoUpstreamQueryResult struct {
	Status     string         `json:"status"`
	Progress   int            `json:"progress"`
	ContentURL *string        `json:"content_url,omitempty"`
	Seconds    *int           `json:"seconds,omitempty"`
	Response   map[string]any `json:"response"`
	ErrorCode  string         `json:"error_code,omitempty"`
	ErrorMsg   string         `json:"error_msg,omitempty"`
}

type VideoUpstreamError struct {
	StatusCode int
	Message    string
	Temporary  bool
}

func (e *VideoUpstreamError) Error() string {
	if e == nil {
		return ""
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("video upstream http %d: %s", e.StatusCode, e.Message)
	}
	return "video upstream request failed: " + e.Message
}

func IsRetryableVideoError(err error) bool {
	if err == nil {
		return false
	}
	var upstreamErr *VideoUpstreamError
	if errors.As(err, &upstreamErr) {
		return upstreamErr.Temporary
	}
	return false
}

type HTTPVideoUpstreamClient struct {
	client *http.Client
	cfg    *config.Config
}

func NewHTTPVideoUpstreamClient(cfg ...*config.Config) VideoUpstreamClient {
	var c *config.Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	return &HTTPVideoUpstreamClient{client: newVideoHTTPClient(c, 60*time.Second), cfg: c}
}

func (c *HTTPVideoUpstreamClient) Create(ctx context.Context, account *Account, template *VideoCallTemplate, body map[string]any) (*VideoUpstreamCreateResult, error) {
	respMap, raw, err := c.doJSON(ctx, account, template.CreateMethod, template.CreatePath, "", body, timeoutSeconds(template.TimeoutConfig, "create_seconds", 60))
	if err != nil {
		return nil, err
	}
	taskID := strings.TrimSpace(firstGJSON(raw, "id", "task_id", "data.id", "data.task_id", "data.video_id"))
	if taskID == "" {
		return nil, fmt.Errorf("%w: upstream task id not found", ErrVideoInvalidRequest)
	}
	return &VideoUpstreamCreateResult{TaskID: taskID, Response: respMap}, nil
}

func (c *HTTPVideoUpstreamClient) Query(ctx context.Context, account *Account, template *VideoCallTemplate, upstreamTaskID string) (*VideoUpstreamQueryResult, error) {
	respMap, raw, err := c.doJSON(ctx, account, template.QueryMethod, template.QueryPath, upstreamTaskID, nil, timeoutSeconds(template.TimeoutConfig, "query_seconds", 30))
	if err != nil {
		return nil, err
	}
	upstreamStatus := strings.ToLower(strings.TrimSpace(firstGJSON(raw, "status", "data.status", "state", "data.state")))
	status := template.StatusMapping[upstreamStatus]
	if status == "" {
		status = VideoStatusInProgress
	}
	contentPath := strings.TrimSpace(template.ResultMapping["content_url"])
	if contentPath == "" {
		contentPath = "data.video_url"
	}
	content := strings.TrimSpace(gjson.GetBytes(raw, contentPath).String())
	progress := int(gjson.GetBytes(raw, fallbackPath(template.ResultMapping["progress"], "progress")).Int())
	if progress <= 0 && status == VideoStatusCompleted {
		progress = 100
	}
	var seconds *int
	secondsPath := strings.TrimSpace(template.ResultMapping["seconds"])
	if secondsPath != "" {
		if v := int(gjson.GetBytes(raw, secondsPath).Int()); v > 0 {
			seconds = &v
		}
	}
	code := strings.TrimSpace(gjson.GetBytes(raw, fallbackPath(template.ErrorMapping["code"], "error.code")).String())
	msg := strings.TrimSpace(gjson.GetBytes(raw, fallbackPath(template.ErrorMapping["message"], "error.message")).String())
	result := &VideoUpstreamQueryResult{
		Status:    status,
		Progress:  progress,
		Seconds:   seconds,
		Response:  respMap,
		ErrorCode: code,
		ErrorMsg:  msg,
	}
	if content != "" {
		result.ContentURL = &content
	}
	return result, nil
}

func (c *HTTPVideoUpstreamClient) Cancel(ctx context.Context, account *Account, template *VideoCallTemplate, upstreamTaskID string) error {
	if template.CancelMethod == nil || template.CancelPath == nil || strings.TrimSpace(*template.CancelPath) == "" {
		return nil
	}
	_, _, err := c.doJSON(ctx, account, *template.CancelMethod, *template.CancelPath, upstreamTaskID, nil, timeoutSeconds(template.TimeoutConfig, "query_seconds", 30))
	return err
}

const videoTemplateRecognitionPrompt = `You convert an upstream asynchronous video-generation API documentation into a JSON call template for a gateway.

Return ONLY a single minified JSON object, no markdown, no prose, no code fences. The object MUST have exactly these keys:
{
  "name": string,                       // a short human-readable template name
  "create_method": "POST"|"GET"|...,    // HTTP method to create a task
  "create_path": string,                // path relative to base_url, e.g. "/v1/videos"
  "query_method": "GET"|...,            // HTTP method to query a task
  "query_path": string,                 // use "{task_id}" placeholder, e.g. "/v1/videos/{task_id}"
  "content_method": string|null,        // method to download the result, null if none
  "content_path": string|null,          // path to download, "{task_id}" allowed, null if result URL comes from query
  "cancel_method": string|null,         // method to cancel, null if unsupported
  "cancel_path": string|null,           // path to cancel, "{task_id}" allowed, null if unsupported
  "status_mapping": { "<upstream_status>": "queued|in_progress|completed|failed|cancelled|expired" },
  "result_mapping": { "content_url": "<gjson path>", "seconds": "<gjson path>", "progress": "<gjson path>" },
  "error_mapping": { "code": "<gjson path>", "message": "<gjson path>" },
  "poll_config": { "interval_seconds": number, "backoff_max_seconds": number, "max_attempts": number },
  "timeout_config": { "create_seconds": number, "query_seconds": number, "content_seconds": number }
}

Rules:
- Paths are relative to the account base_url; do NOT include the host.
- Use the literal token {task_id} wherever the upstream task id goes in query/content/cancel paths.
- result_mapping/error_mapping values are gjson paths into the QUERY response body (e.g. "data.video_url", "data.task_result.videos.0.url"). Not JSONPath.
- status_mapping keys are the raw upstream status strings (lowercase), values are one of the six unified statuses listed above.
- If the upstream returns the video URL directly in the query response (no separate download endpoint), set content_method and content_path to null.
- If a field is unknown, use sensible defaults: poll interval 5, backoff 30, max_attempts 240; timeouts create 60, query 30, content 300.`

func (c *HTTPVideoUpstreamClient) RecognizeTemplate(ctx context.Context, account *Account, model, doc string) (*VideoCallTemplate, error) {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return nil, fmt.Errorf("%w: document is required", ErrVideoInvalidRequest)
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("%w: model is required", ErrVideoInvalidRequest)
	}
	body := map[string]any{
		"model":       model,
		"temperature": 0,
		"messages": []map[string]any{
			{"role": "system", "content": videoTemplateRecognitionPrompt},
			{"role": "user", "content": "Upstream API documentation:\n\n" + doc},
		},
	}
	_, raw, err := c.doJSON(ctx, account, http.MethodPost, "/v1/chat/completions", "", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(gjson.GetBytes(raw, "choices.0.message.content").String())
	if content == "" {
		return nil, fmt.Errorf("%w: AI returned empty content", ErrVideoInvalidRequest)
	}
	jsonText := extractJSONObject(content)
	if jsonText == "" {
		return nil, fmt.Errorf("%w: AI response did not contain a JSON object", ErrVideoInvalidRequest)
	}
	template, err := parseRecognizedTemplate(jsonText)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// extractJSONObject pulls the first balanced top-level JSON object out of text,
// tolerating ```json fences or leading/trailing prose.
func extractJSONObject(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inStr {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inStr = false
			}
			continue
		}
		switch ch {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}

func parseRecognizedTemplate(jsonText string) (*VideoCallTemplate, error) {
	var dto struct {
		Name          string            `json:"name"`
		CreateMethod  string            `json:"create_method"`
		CreatePath    string            `json:"create_path"`
		QueryMethod   string            `json:"query_method"`
		QueryPath     string            `json:"query_path"`
		ContentMethod *string           `json:"content_method"`
		ContentPath   *string           `json:"content_path"`
		CancelMethod  *string           `json:"cancel_method"`
		CancelPath    *string           `json:"cancel_path"`
		StatusMapping map[string]string `json:"status_mapping"`
		ResultMapping map[string]string `json:"result_mapping"`
		ErrorMapping  map[string]string `json:"error_mapping"`
		PollConfig    map[string]any    `json:"poll_config"`
		TimeoutConfig map[string]any    `json:"timeout_config"`
	}
	if err := json.Unmarshal([]byte(jsonText), &dto); err != nil {
		return nil, fmt.Errorf("%w: AI returned invalid JSON: %s", ErrVideoInvalidRequest, err.Error())
	}
	return &VideoCallTemplate{
		Name:          strings.TrimSpace(dto.Name),
		CreateMethod:  dto.CreateMethod,
		CreatePath:    dto.CreatePath,
		QueryMethod:   dto.QueryMethod,
		QueryPath:     dto.QueryPath,
		ContentMethod: trimOptionalString(dto.ContentMethod),
		ContentPath:   trimOptionalString(dto.ContentPath),
		CancelMethod:  trimOptionalString(dto.CancelMethod),
		CancelPath:    trimOptionalString(dto.CancelPath),
		StatusMapping: dto.StatusMapping,
		ResultMapping: dto.ResultMapping,
		ErrorMapping:  dto.ErrorMapping,
		PollConfig:    dto.PollConfig,
		TimeoutConfig: dto.TimeoutConfig,
		Status:        "active",
	}, nil
}

func trimOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (c *HTTPVideoUpstreamClient) doJSON(ctx context.Context, account *Account, method, path, taskID string, body map[string]any, timeout time.Duration) (map[string]any, []byte, error) {
	if c.client == nil {
		c.client = newVideoHTTPClient(c.cfg, timeout)
	}
	baseURL := videoAccountBaseURL(account)
	if baseURL == "" {
		return nil, nil, fmt.Errorf("%w: account base_url is required", ErrVideoInvalidRequest)
	}
	endpoint, err := joinVideoURL(baseURL, strings.ReplaceAll(path, "{task_id}", url.PathEscape(taskID)))
	if err != nil {
		return nil, nil, err
	}
	if normalized, err := normalizeVideoOutboundURL(endpoint, c.cfg); err != nil {
		return nil, nil, fmt.Errorf("%w: invalid upstream url: %s", ErrVideoInvalidRequest, err.Error())
	} else {
		endpoint = normalized
	}
	var reader io.Reader
	if body != nil {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, strings.ToUpper(defaultString(method, http.MethodGet)), endpoint, reader)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	applyVideoAuth(req, account)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, &VideoUpstreamError{Message: err.Error(), Temporary: true}
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if readErr != nil {
		return nil, nil, &VideoUpstreamError{StatusCode: resp.StatusCode, Message: readErr.Error(), Temporary: true}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(firstGJSON(raw, "error.message", "message", "data.message"))
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return nil, raw, &VideoUpstreamError{
			StatusCode: resp.StatusCode,
			Message:    msg,
			Temporary:  resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode >= 500,
		}
	}
	var respMap map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &respMap); err != nil {
			return nil, raw, fmt.Errorf("video upstream returned non-json response: %w", err)
		}
	}
	if respMap == nil {
		respMap = map[string]any{}
	}
	return respMap, raw, nil
}

func videoAccountBaseURL(account *Account) string {
	if account == nil {
		return ""
	}
	if raw := strings.TrimSpace(account.GetCredential("base_url")); raw != "" {
		return strings.TrimRight(raw, "/")
	}
	return strings.TrimRight(account.GetBaseURL(), "/")
}

func applyVideoAuth(req *http.Request, account *Account) {
	if req == nil || account == nil {
		return
	}
	if v := strings.TrimSpace(account.GetCredential("authorization")); v != "" {
		req.Header.Set("Authorization", v)
		return
	}
	if v := strings.TrimSpace(account.GetCredential("access_token")); v != "" {
		req.Header.Set("Authorization", "Bearer "+v)
		return
	}
	if v := strings.TrimSpace(account.GetCredential("api_key")); v != "" {
		req.Header.Set("Authorization", "Bearer "+v)
	}
}

func joinVideoURL(baseURL, path string) (string, error) {
	base, err := url.Parse(strings.TrimRight(baseURL, "/") + "/")
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(strings.TrimLeft(path, "/"))
	if err != nil {
		return "", err
	}
	return base.ResolveReference(ref).String(), nil
}

func timeoutSeconds(cfg map[string]any, key string, fallback int) time.Duration {
	seconds := intFromAny(cfg[key])
	if seconds <= 0 {
		seconds = fallback
	}
	return time.Duration(seconds) * time.Second
}

func firstGJSON(raw []byte, paths ...string) string {
	for _, path := range paths {
		if v := strings.TrimSpace(gjson.GetBytes(raw, path).String()); v != "" {
			return v
		}
	}
	return ""
}

func fallbackPath(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
