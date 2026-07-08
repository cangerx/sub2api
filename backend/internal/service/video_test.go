package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/ccapi/internal/config"
)

func TestVideoValidateURLsUsesConfiguredAllowlist(t *testing.T) {
	svc := &VideoService{cfg: &config.Config{}}
	svc.cfg.Security.URLAllowlist.Enabled = true
	svc.cfg.Security.URLAllowlist.UpstreamHosts = []string{"cdn.example.com", "*.trusted.example.com"}

	if err := svc.ValidateVideoURLs(VideoCreateRequest{
		InputReference: "https://cdn.example.com/assets/frame.png",
		ExtraBody: map[string]any{
			"references": []any{"https://media.trusted.example.com/ref.png"},
		},
	}); err != nil {
		t.Fatalf("expected allowlisted video URLs to pass: %v", err)
	}

	err := svc.ValidateVideoURLs(VideoCreateRequest{
		ExtraBody: map[string]any{
			"references": []any{"https://evil.example.net/ref.png"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "host is not allowed") {
		t.Fatalf("expected non-allowlisted nested URL to fail, got %v", err)
	}
}

func TestReserveVideoSecondsUsesMaxPossibleDuration(t *testing.T) {
	model := &VideoModel{
		Limits:           map[string]any{"max_seconds": 15},
		SupportedOptions: map[string]any{"seconds": []any{5, 10}},
	}
	if got := reserveVideoSeconds(model, 5); got != 15 {
		t.Fatalf("reserveVideoSeconds = %d, want 15", got)
	}

	model.Limits = map[string]any{}
	if got := reserveVideoSeconds(model, 5); got != 10 {
		t.Fatalf("reserveVideoSeconds fallback = %d, want 10", got)
	}

	if got := reserveVideoSeconds(model, 20); got != 20 {
		t.Fatalf("reserveVideoSeconds should not reduce requested duration, got %d", got)
	}
}

func TestRegisteredVideoRequestShapesIsSorted(t *testing.T) {
	got := RegisteredVideoRequestShapes()
	want := []string{"grok_imagine", "seedance", "videos"}
	if len(got) != len(want) {
		t.Fatalf("shape count = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("shape[%d] = %q, want %q; all=%v", i, got[i], want[i], got)
		}
	}
}

func TestVideoUpstreamResultsUseSnakeCaseJSON(t *testing.T) {
	body, err := json.Marshal(VideoUpstreamCreateResult{
		TaskID:   "task_123",
		Response: map[string]any{"ok": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw := string(body)
	if !strings.Contains(raw, `"task_id"`) || strings.Contains(raw, `"TaskID"`) {
		t.Fatalf("unexpected create result json: %s", raw)
	}

	contentURL := "https://cdn.example.com/video.mp4"
	body, err = json.Marshal(VideoUpstreamQueryResult{
		Status:     VideoStatusCompleted,
		Progress:   100,
		ContentURL: &contentURL,
		Response:   map[string]any{"ok": true},
		ErrorCode:  "upstream_failed",
		ErrorMsg:   "failed",
	})
	if err != nil {
		t.Fatal(err)
	}
	raw = string(body)
	for _, field := range []string{`"content_url"`, `"error_code"`, `"error_msg"`} {
		if !strings.Contains(raw, field) {
			t.Fatalf("query result json missing %s: %s", field, raw)
		}
	}
}

func TestHTTPVideoUpstreamQuerySupportsContentURLFallbacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/videos/task_123" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"task_id": "task_123",
			"status": "completed",
			"progress": 100,
			"seconds": "5",
			"metadata": {
				"content_url": "https://cdn.example.com/video.mp4"
			}
		}`))
	}))
	defer server.Close()

	client := NewHTTPVideoUpstreamClient().(*HTTPVideoUpstreamClient)
	result, err := client.Query(context.Background(), &Account{
		Credentials: map[string]any{
			"base_url": server.URL,
			"api_key":  "test-key",
		},
	}, &VideoCallTemplate{
		QueryMethod:   "GET",
		QueryPath:     "/v1/videos/{task_id}",
		StatusMapping: map[string]string{"completed": VideoStatusCompleted},
		ResultMapping: map[string]string{
			"content_url": "video_url|url|metadata.content_url",
			"seconds":     "seconds",
			"progress":    "progress",
		},
	}, "task_123")
	if err != nil {
		t.Fatalf("query upstream: %v", err)
	}
	if result.ContentURL == nil || *result.ContentURL != "https://cdn.example.com/video.mp4" {
		t.Fatalf("content_url = %v, want metadata.content_url fallback", result.ContentURL)
	}
	if result.Seconds == nil || *result.Seconds != 5 {
		t.Fatalf("seconds = %v, want 5", result.Seconds)
	}
}

func TestBuildVideoUpstreamRequestUsesMappedUpstreamModel(t *testing.T) {
	model := &VideoModel{
		PublicModel:      "video-fast",
		UpstreamModelID:  "legacy-should-not-be-used",
		RequestShape:     "videos",
		Defaults:         map[string]any{"seconds": "10", "size": "1280x720"},
		SupportedOptions: map[string]any{},
	}
	req := &VideoCreateRequest{Model: "video-fast", Prompt: "a cat", ExtraBody: map[string]any{}}

	body, err := BuildVideoUpstreamRequest(req, model, "videos-fast")
	if err != nil {
		t.Fatalf("build upstream request: %v", err)
	}
	if got := body["model"]; got != "videos-fast" {
		t.Fatalf("upstream body model = %v, want videos-fast (mapped, not video_models.upstream_model_id)", got)
	}
}

func TestVideoModelObjectExposesExtraBodyAllow(t *testing.T) {
	body, err := json.Marshal(VideoModelObject{
		ID:             "seedance",
		Object:         "model",
		Status:         "active",
		ExtraBodyAllow: []string{"seed", "camera"},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw := string(body)
	if !strings.Contains(raw, `"extra_body_allow"`) || strings.Contains(raw, `"ExtraBodyAllow"`) {
		t.Fatalf("unexpected model json: %s", raw)
	}
}

func TestVideoCandidatePlatformsIncludesExplicitVideoScopes(t *testing.T) {
	got := videoCandidatePlatforms(&Group{
		SupportedModelScopes: []string{"claude", "video:grok", " video:custom ", "video:grok"},
	})
	want := []string{PlatformVideo, "grok", "custom"}
	if len(got) != len(want) {
		t.Fatalf("platform count = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("platform[%d] = %q, want %q; all=%v", i, got[i], want[i], got)
		}
	}
}

func TestFilterVideoCapableAccountsRequiresCapabilityAndModelSupport(t *testing.T) {
	accounts := []Account{
		{ID: 1, Platform: PlatformVideo},
		{
			ID:          2,
			Platform:    "grok",
			Credentials: map[string]any{"capabilities": []any{"videos"}, "model_mapping": map[string]any{"grok-video": "grok-video"}},
		},
		{
			ID:          3,
			Platform:    PlatformOpenAI,
			Credentials: map[string]any{"capabilities": []any{"chat"}},
		},
		{
			ID:          4,
			Platform:    "custom",
			Credentials: map[string]any{"capabilities": []any{"videos"}, "model_mapping": map[string]any{"other-video": "other-video"}},
		},
	}

	got := filterVideoCapableAccounts(accounts, "grok-video")
	if len(got) != 2 {
		t.Fatalf("filtered account count = %d, want 2: %+v", len(got), got)
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("filtered account IDs = [%d %d], want [1 2]", got[0].ID, got[1].ID)
	}
}

func TestVideoServiceSelectVideoAccountCanUseExplicitCapabilityPlatform(t *testing.T) {
	repo := &videoCapabilityAccountRepo{
		accounts: []Account{
			{
				ID:          42,
				Platform:    "grok",
				Status:      StatusActive,
				Schedulable: true,
				Priority:    1,
				Credentials: map[string]any{
					"capabilities":  []any{"videos"},
					"model_mapping": map[string]any{"grok-video": "grok-video-upstream"},
				},
			},
		},
	}
	svc := &VideoService{accountRepo: repo}
	groupID := int64(7)
	account, err := svc.selectVideoAccount(context.Background(), &groupID, &Group{
		ID:                   groupID,
		SupportedModelScopes: []string{"video:grok"},
	}, "grok-video")
	if err != nil {
		t.Fatalf("selectVideoAccount returned error: %v", err)
	}
	if account == nil || account.ID != 42 {
		t.Fatalf("selected account = %+v, want id 42", account)
	}
	if len(repo.requestedPlatforms) != 2 || repo.requestedPlatforms[0] != PlatformVideo || repo.requestedPlatforms[1] != "grok" {
		t.Fatalf("requested platforms = %v, want [video grok]", repo.requestedPlatforms)
	}
}

func TestAdminRecognizeTemplateRejectsNonVideoAccount(t *testing.T) {
	upstream := &recordingVideoUpstreamClient{}
	svc := &VideoService{
		accountRepo: &videoCapabilityAccountRepo{
			accounts: []Account{{ID: 7, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive}},
		},
		upstreamClient: upstream,
	}

	_, err := svc.AdminRecognizeTemplate(context.Background(), 7, "gpt-4o-mini", "POST /v1/videos")
	if err == nil {
		t.Fatal("expected non-video account to be rejected")
	}
	if !errors.Is(err, ErrVideoInvalidRequest) {
		t.Fatalf("expected ErrVideoInvalidRequest, got %v", err)
	}
	if !strings.Contains(err.Error(), "video platform account") {
		t.Fatalf("expected video platform account error, got %v", err)
	}
	if upstream.recognizeCalls != 0 {
		t.Fatalf("upstream RecognizeTemplate was called %d times, want 0", upstream.recognizeCalls)
	}
}

func TestAdminRecognizeTemplateRejectsInactiveVideoAccount(t *testing.T) {
	upstream := &recordingVideoUpstreamClient{}
	svc := &VideoService{
		accountRepo: &videoCapabilityAccountRepo{
			accounts: []Account{{ID: 9, Platform: PlatformVideo, Type: AccountTypeAPIKey, Status: StatusDisabled}},
		},
		upstreamClient: upstream,
	}

	_, err := svc.AdminRecognizeTemplate(context.Background(), 9, "video-chat-model", "POST /v1/videos")
	if err == nil {
		t.Fatal("expected inactive video account to be rejected")
	}
	if !errors.Is(err, ErrVideoInvalidRequest) {
		t.Fatalf("expected ErrVideoInvalidRequest, got %v", err)
	}
	if !strings.Contains(err.Error(), "account must be active") {
		t.Fatalf("expected active account error, got %v", err)
	}
	if upstream.recognizeCalls != 0 {
		t.Fatalf("upstream RecognizeTemplate was called %d times, want 0", upstream.recognizeCalls)
	}
}

func TestAdminRecognizeTemplateRejectsNonAPIKeyVideoAccount(t *testing.T) {
	upstream := &recordingVideoUpstreamClient{}
	svc := &VideoService{
		accountRepo: &videoCapabilityAccountRepo{
			accounts: []Account{{ID: 10, Platform: PlatformVideo, Type: AccountTypeOAuth, Status: StatusActive}},
		},
		upstreamClient: upstream,
	}

	_, err := svc.AdminRecognizeTemplate(context.Background(), 10, "video-chat-model", "POST /v1/videos")
	if err == nil {
		t.Fatal("expected non-API-key video account to be rejected")
	}
	if !errors.Is(err, ErrVideoInvalidRequest) {
		t.Fatalf("expected ErrVideoInvalidRequest, got %v", err)
	}
	if !strings.Contains(err.Error(), "video API key account") {
		t.Fatalf("expected video API key account error, got %v", err)
	}
	if upstream.recognizeCalls != 0 {
		t.Fatalf("upstream RecognizeTemplate was called %d times, want 0", upstream.recognizeCalls)
	}
}

func TestAdminRecognizeTemplateAcceptsVideoAccount(t *testing.T) {
	upstream := &recordingVideoUpstreamClient{
		template: &VideoCallTemplate{
			Name:          "Video API",
			CreateMethod:  "POST",
			CreatePath:    "/v1/videos",
			QueryMethod:   "GET",
			QueryPath:     "/v1/videos/{task_id}",
			StatusMapping: map[string]string{"completed": VideoStatusCompleted},
		},
	}
	svc := &VideoService{
		accountRepo: &videoCapabilityAccountRepo{
			accounts: []Account{{ID: 8, Platform: PlatformVideo, Type: AccountTypeAPIKey, Status: StatusActive}},
		},
		upstreamClient: upstream,
	}

	template, err := svc.AdminRecognizeTemplate(context.Background(), 8, "video-chat-model", "POST /v1/videos")
	if err != nil {
		t.Fatalf("expected video account to be accepted: %v", err)
	}
	if template == nil || template.CreatePath != "/v1/videos" {
		t.Fatalf("unexpected template: %+v", template)
	}
	if upstream.recognizeCalls != 1 {
		t.Fatalf("upstream RecognizeTemplate calls = %d, want 1", upstream.recognizeCalls)
	}
	if upstream.account == nil || upstream.account.ID != 8 {
		t.Fatalf("upstream account = %+v, want id 8", upstream.account)
	}
}

type videoCapabilityAccountRepo struct {
	AccountRepository
	accounts           []Account
	requestedPlatforms []string
}

func (r *videoCapabilityAccountRepo) ListSchedulableByGroupIDAndPlatforms(_ context.Context, _ int64, platforms []string) ([]Account, error) {
	r.requestedPlatforms = append([]string(nil), platforms...)
	return filterAccountsByPlatforms(r.accounts, platforms), nil
}

func (r *videoCapabilityAccountRepo) ListSchedulableUngroupedByPlatforms(_ context.Context, platforms []string) ([]Account, error) {
	r.requestedPlatforms = append([]string(nil), platforms...)
	return filterAccountsByPlatforms(r.accounts, platforms), nil
}

func (r *videoCapabilityAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	for i := range r.accounts {
		if r.accounts[i].ID == id {
			return &r.accounts[i], nil
		}
	}
	return nil, ErrAccountNotFound
}

func filterAccountsByPlatforms(accounts []Account, platforms []string) []Account {
	allowed := make(map[string]struct{}, len(platforms))
	for _, platform := range platforms {
		allowed[platform] = struct{}{}
	}
	var out []Account
	for _, account := range accounts {
		if _, ok := allowed[account.Platform]; ok {
			out = append(out, account)
		}
	}
	return out
}

type recordingVideoUpstreamClient struct {
	VideoUpstreamClient
	recognizeCalls int
	account        *Account
	template       *VideoCallTemplate
	err            error
}

func (c *recordingVideoUpstreamClient) RecognizeTemplate(_ context.Context, account *Account, _, _ string) (*VideoCallTemplate, error) {
	c.recognizeCalls++
	c.account = account
	if c.err != nil {
		return nil, c.err
	}
	return c.template, nil
}
