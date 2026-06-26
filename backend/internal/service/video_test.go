package service

import (
	"context"
	"encoding/json"
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
