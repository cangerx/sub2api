package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/ccapi/internal/config"
	"github.com/Wei-Shaw/ccapi/internal/util/urlvalidator"
	"github.com/google/uuid"
)

const (
	VideoStatusQueued     = "queued"
	VideoStatusInProgress = "in_progress"
	VideoStatusCompleted  = "completed"
	VideoStatusFailed     = "failed"
	VideoStatusCancelled  = "cancelled"
	VideoStatusExpired    = "expired"

	VideoBillingStateReserved = "reserved"
	VideoBillingStateSettled  = "settled"

	VideoObjectType = "video"
)

var (
	ErrVideoModelNotFound       = errors.New("video model not found")
	ErrVideoModelDisabled       = errors.New("video model disabled")
	ErrVideoTemplateNotFound    = errors.New("video call template not found")
	ErrVideoTaskNotFound        = errors.New("video task not found")
	ErrVideoInvalidRequest      = errors.New("invalid video request")
	ErrVideoPricingUnavailable  = errors.New("video pricing unavailable")
	ErrVideoInsufficientBalance = errors.New("insufficient balance for video task")
	ErrVideoContentUnavailable  = errors.New("video content unavailable")
	ErrVideoContentExpired      = errors.New("video content expired")
)

var videoContentHTTPClient = &http.Client{
	Timeout: 10 * time.Minute,
	Transport: &http.Transport{
		Proxy:       http.ProxyFromEnvironment,
		DialContext: safeDialContext,
	},
}

type VideoCallTemplate struct {
	ID            int64
	Name          string
	CreateMethod  string
	CreatePath    string
	QueryMethod   string
	QueryPath     string
	ContentMethod *string
	ContentPath   *string
	CancelMethod  *string
	CancelPath    *string
	StatusMapping map[string]string
	ResultMapping map[string]string
	ErrorMapping  map[string]string
	PollConfig    map[string]any
	TimeoutConfig map[string]any
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type VideoModel struct {
	ID               int64
	PublicModel      string
	DisplayName      *string
	TemplateID       int64
	UpstreamModelID  string
	RequestShape     string
	Status           string
	Capabilities     map[string]any
	Defaults         map[string]any
	Limits           map[string]any
	SupportedOptions map[string]any
	ExtraBodyAllow   []string
	SortOrder        int
	Template         *VideoCallTemplate
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type VideoGenerationTask struct {
	ID                      int64
	PublicID                string
	UserID                  int64
	APIKeyID                int64
	GroupID                 *int64
	AccountID               int64
	ChannelID               *int64
	VideoModelID            int64
	RequestedModel          string
	UpstreamModel           string
	UpstreamTaskID          *string
	Status                  string
	Progress                int
	BillingState            string
	RequestPayload          map[string]any
	UpstreamRequestPayload  map[string]any
	UpstreamResponsePayload map[string]any
	ResultPayload           map[string]any
	ErrorCode               *string
	ErrorMessage            *string
	ContentURL              *string
	UpstreamContentURL      *string
	LocalContentURL         *string
	BillingMode             string
	UnitPrice               float64
	UnitSeconds             *float64
	RequestedSeconds        *int
	BillableSeconds         *int
	ReservedCost            float64
	EstimatedCost           float64
	ActualCost              float64
	IdempotencyKey          *string
	SubmittedAt             *time.Time
	StartedAt               *time.Time
	CompletedAt             *time.Time
	ExpiresAt               *time.Time
	NextPollAt              *time.Time
	PollCount               int
	LockedUntil             *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time

	VideoModel *VideoModel
}

type VideoCreateRequest struct {
	Model          string         `json:"model"`
	Prompt         string         `json:"prompt"`
	Seconds        string         `json:"seconds"`
	Size           string         `json:"size"`
	InputReference string         `json:"input_reference"`
	ExtraBody      map[string]any `json:"extra_body"`
	Raw            map[string]any `json:"-"`
}

type VideoCreateInput struct {
	Request        VideoCreateRequest
	User           *User
	APIKey         *APIKey
	GroupID        *int64
	IdempotencyKey string
	UserAgent      string
	IPAddress      string
}

type VideoObject struct {
	ID          string            `json:"id"`
	Object      string            `json:"object"`
	Model       string            `json:"model"`
	Status      string            `json:"status"`
	Progress    int               `json:"progress"`
	CreatedAt   int64             `json:"created_at"`
	CompletedAt *int64            `json:"completed_at"`
	ExpiresAt   *int64            `json:"expires_at"`
	Seconds     string            `json:"seconds,omitempty"`
	Size        string            `json:"size,omitempty"`
	Error       *VideoObjectError `json:"error"`
}

type VideoObjectError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type VideoListResponse struct {
	Object string        `json:"object"`
	Data   []VideoObject `json:"data"`
}

type VideoModelListResponse struct {
	Object string             `json:"object"`
	Data   []VideoModelObject `json:"data"`
}

type VideoModelObject struct {
	ID             string         `json:"id"`
	Object         string         `json:"object"`
	DisplayName    string         `json:"display_name,omitempty"`
	Status         string         `json:"status"`
	Supports       []string       `json:"supports,omitempty"`
	Seconds        []int          `json:"seconds,omitempty"`
	Sizes          []string       `json:"sizes,omitempty"`
	Limits         map[string]any `json:"limits,omitempty"`
	ExtraBodyAllow []string       `json:"extra_body_allow,omitempty"`
	Billing        *VideoBilling  `json:"billing,omitempty"`
}

type VideoBilling struct {
	Mode        string  `json:"mode"`
	UnitPrice   float64 `json:"unit_price"`
	UnitSeconds float64 `json:"unit_seconds,omitempty"`
	Currency    string  `json:"currency"`
}

type VideoRepository interface {
	GetModelByPublicModel(ctx context.Context, model string) (*VideoModel, error)
	ListActiveModels(ctx context.Context) ([]VideoModel, error)
	GetTemplateByID(ctx context.Context, id int64) (*VideoCallTemplate, error)
	ListTemplates(ctx context.Context) ([]VideoCallTemplate, error)
	CreateTemplate(ctx context.Context, template *VideoCallTemplate) error
	UpdateTemplate(ctx context.Context, template *VideoCallTemplate) error
	DeleteTemplate(ctx context.Context, id int64) error
	ListModels(ctx context.Context, includeDisabled bool) ([]VideoModel, error)
	GetModelByID(ctx context.Context, id int64) (*VideoModel, error)
	CreateModel(ctx context.Context, model *VideoModel) error
	UpdateModel(ctx context.Context, model *VideoModel) error
	DeleteModel(ctx context.Context, id int64) error
	CreateTask(ctx context.Context, task *VideoGenerationTask) error
	GetTaskByPublicID(ctx context.Context, publicID string) (*VideoGenerationTask, error)
	GetTaskByIdempotencyKey(ctx context.Context, key string) (*VideoGenerationTask, error)
	ListTasksByAPIKey(ctx context.Context, apiKeyID int64, limit int, after string) ([]VideoGenerationTask, error)
	ListTasks(ctx context.Context, filter VideoTaskFilter) ([]VideoGenerationTask, int64, error)
	ClaimDueTasks(ctx context.Context, limit int, lockFor time.Duration) ([]VideoGenerationTask, error)
	MarkTaskSubmitted(ctx context.Context, publicID, upstreamTaskID string, upstreamRequest, upstreamResponse map[string]any, nextPollAt time.Time) error
	MarkTaskPollResult(ctx context.Context, publicID string, status string, progress int, upstreamResponse map[string]any, contentURL *string, resultPayload map[string]any, nextPollAt *time.Time) error
	MarkTaskFailed(ctx context.Context, publicID, code, message string) error
	MarkTaskFailedNoSettlement(ctx context.Context, publicID, code, message string) error
	MarkTaskCompleted(ctx context.Context, publicID string, progress int, contentURL *string, resultPayload map[string]any, billableSeconds *int, actualCost float64, expiresAt time.Time) error
	MarkTaskCancelled(ctx context.Context, publicID string) (bool, error)
	MarkTaskExpired(ctx context.Context, publicID string) error
	RequeueTask(ctx context.Context, publicID string, nextPollAt time.Time) error
	ScheduleTaskRetry(ctx context.Context, publicID string, nextPollAt time.Time, code, message string) error
}

type VideoTaskFilter struct {
	Status   string
	Model    string
	UserID   int64
	APIKeyID int64
	StartAt  *time.Time
	EndAt    *time.Time
	Limit    int
	Offset   int
}

type VideoService struct {
	repo            VideoRepository
	accountRepo     AccountRepository
	userRepo        UserRepository
	usageBilling    UsageBillingRepository
	usageLogRepo    UsageLogRepository
	channelService  *ChannelService
	pricingResolver *ModelPricingResolver
	concurrency     *ConcurrencyService
	cfg             *config.Config
	upstreamClient  VideoUpstreamClient
}

func NewVideoService(repo VideoRepository, accountRepo AccountRepository, userRepo UserRepository, usageBilling UsageBillingRepository, usageLogRepo UsageLogRepository, channelService *ChannelService, pricingResolver *ModelPricingResolver, concurrency *ConcurrencyService, cfg *config.Config) *VideoService {
	return &VideoService{
		repo:            repo,
		accountRepo:     accountRepo,
		userRepo:        userRepo,
		usageBilling:    usageBilling,
		usageLogRepo:    usageLogRepo,
		channelService:  channelService,
		pricingResolver: pricingResolver,
		concurrency:     concurrency,
		cfg:             cfg,
		upstreamClient:  NewHTTPVideoUpstreamClient(cfg),
	}
}

func (s *VideoService) SetUpstreamClient(client VideoUpstreamClient) {
	if client != nil {
		s.upstreamClient = client
	}
}

func (s *VideoService) Create(ctx context.Context, input VideoCreateInput) (*VideoGenerationTask, error) {
	if input.APIKey == nil || input.User == nil {
		return nil, fmt.Errorf("%w: missing auth context", ErrVideoInvalidRequest)
	}
	req := input.Request
	req.Model = strings.TrimSpace(req.Model)
	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Model == "" {
		return nil, fmt.Errorf("%w: model is required", ErrVideoInvalidRequest)
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("%w: prompt is required", ErrVideoInvalidRequest)
	}
	if err := s.ValidateVideoURLs(req); err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.IdempotencyKey) != "" && s.repo != nil {
		key := videoIdempotencyHash(input.APIKey.ID, input.IdempotencyKey)
		existing, err := s.repo.GetTaskByIdempotencyKey(ctx, key)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	model, err := s.repo.GetModelByPublicModel(ctx, req.Model)
	if err != nil {
		return nil, err
	}
	if model.Status == "disabled" {
		return nil, ErrVideoModelDisabled
	}
	if model.Status == "" {
		model.Status = "active"
	}

	seconds, err := resolveVideoSeconds(req, model)
	if err != nil {
		return nil, err
	}
	size, _, _, err := resolveVideoSize(req, model)
	if err != nil {
		return nil, err
	}

	account, err := s.selectVideoAccount(ctx, input.GroupID, input.APIKey.Group, model.PublicModel)
	if err != nil {
		return nil, err
	}

	// Upstream model name comes from the account-level model mapping (public
	// model → upstream model), reusing the same mapping other platforms use.
	// No mapping configured ⇒ pass the requested model through unchanged.
	upstreamModel := model.PublicModel
	if mapped, matched := account.ResolveMappedModel(model.PublicModel); matched && strings.TrimSpace(mapped) != "" {
		upstreamModel = mapped
	} else if strings.TrimSpace(model.UpstreamModelID) != "" {
		// Backward-compat fallback for models that still carry an explicit
		// upstream_model_id (legacy rows / no account mapping yet).
		upstreamModel = model.UpstreamModelID
	}

	upstreamBody, err := BuildVideoUpstreamRequest(&req, model, upstreamModel)
	if err != nil {
		return nil, err
	}

	reserveSeconds := reserveVideoSeconds(model, seconds)
	pricing, err := s.resolveVideoPricing(ctx, input.GroupID, model.PublicModel, reserveSeconds)
	if err != nil {
		return nil, err
	}
	if input.User.Balance < pricing.EstimatedCost {
		return nil, ErrVideoInsufficientBalance
	}

	publicID := "video_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	idempotencyKey := ""
	if strings.TrimSpace(input.IdempotencyKey) != "" {
		idempotencyKey = videoIdempotencyHash(input.APIKey.ID, input.IdempotencyKey)
	}
	task := &VideoGenerationTask{
		PublicID:               publicID,
		UserID:                 input.User.ID,
		APIKeyID:               input.APIKey.ID,
		GroupID:                input.GroupID,
		AccountID:              account.ID,
		VideoModelID:           model.ID,
		RequestedModel:         req.Model,
		UpstreamModel:          upstreamModel,
		Status:                 VideoStatusQueued,
		Progress:               0,
		BillingState:           VideoBillingStateReserved,
		RequestPayload:         req.Raw,
		UpstreamRequestPayload: upstreamBody,
		BillingMode:            string(pricing.Mode),
		UnitPrice:              pricing.UnitPrice,
		UnitSeconds:            pricing.UnitSeconds,
		RequestedSeconds:       &seconds,
		ReservedCost:           pricing.EstimatedCost,
		EstimatedCost:          pricing.EstimatedCost,
		ActualCost:             0,
		IdempotencyKey:         nilIfEmpty(idempotencyKey),
		NextPollAt:             videoPtrTime(time.Now().UTC()),
		VideoModel:             model,
	}
	if size != "" {
		task.RequestPayload["size"] = size
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	slot, err := s.acquireVideoSlots(ctx, task, input, account)
	if err != nil {
		_ = s.repo.MarkTaskFailedNoSettlement(ctx, task.PublicID, "concurrency_failed", err.Error())
		return nil, err
	}
	if slot != nil && !slot.Acquired {
		_ = s.repo.MarkTaskFailedNoSettlement(ctx, task.PublicID, "concurrency_limited", "video concurrency limit exceeded")
		return nil, ErrNoAvailableAccounts
	}
	releaseOnCreateFailure := true
	defer func() {
		if releaseOnCreateFailure && slot != nil && slot.ReleaseFunc != nil {
			slot.ReleaseFunc()
		}
	}()
	if err := s.reserveTaskCost(ctx, task, input); err != nil {
		_ = s.repo.MarkTaskFailedNoSettlement(ctx, task.PublicID, "billing_failed", err.Error())
		return nil, err
	}
	if err := s.submitTask(ctx, task); err == nil {
		releaseOnCreateFailure = false
		submitted, getErr := s.repo.GetTaskByPublicID(ctx, task.PublicID)
		if getErr == nil && submitted != nil {
			return submitted, nil
		}
		task.Status = VideoStatusInProgress
		task.Progress = 1
	} else if IsRetryableVideoError(err) {
		releaseOnCreateFailure = false
		task.NextPollAt = videoPtrTime(time.Now().UTC().Add(5 * time.Second))
	} else {
		_ = s.repo.MarkTaskFailed(ctx, task.PublicID, "upstream_create_failed", err.Error())
		return nil, err
	}
	return task, nil
}

func (s *VideoService) Get(ctx context.Context, publicID string, apiKeyID int64) (*VideoGenerationTask, error) {
	task, err := s.repo.GetTaskByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	if task.APIKeyID != apiKeyID {
		return nil, ErrVideoTaskNotFound
	}
	s.applyExpiredState(ctx, task)
	return task, nil
}

func (s *VideoService) List(ctx context.Context, apiKeyID int64, limit int, after string) ([]VideoGenerationTask, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	tasks, err := s.repo.ListTasksByAPIKey(ctx, apiKeyID, limit, after)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		s.applyExpiredState(ctx, &tasks[i])
	}
	return tasks, nil
}

func (s *VideoService) Cancel(ctx context.Context, publicID string, apiKeyID int64) (*VideoGenerationTask, error) {
	task, err := s.Get(ctx, publicID, apiKeyID)
	if err != nil {
		return nil, err
	}
	ok, err := s.repo.MarkTaskCancelled(ctx, publicID)
	if err != nil {
		return nil, err
	}
	if ok {
		task.Status = VideoStatusCancelled
		_ = s.cancelUpstreamTask(ctx, task)
		s.releaseVideoSlots(task)
	}
	return task, nil
}

func (s *VideoService) cancelUpstreamTask(ctx context.Context, task *VideoGenerationTask) error {
	if s.upstreamClient == nil || task == nil || task.UpstreamTaskID == nil || strings.TrimSpace(*task.UpstreamTaskID) == "" {
		return nil
	}
	account, err := s.accountRepo.GetByID(ctx, task.AccountID)
	if err != nil {
		return err
	}
	template := taskTemplate(task)
	if template == nil {
		return ErrVideoTemplateNotFound
	}
	return s.upstreamClient.Cancel(ctx, account, template, *task.UpstreamTaskID)
}

func (s *VideoService) ContentLocation(ctx context.Context, publicID string, apiKeyID int64) (string, error) {
	task, err := s.Get(ctx, publicID, apiKeyID)
	if err != nil {
		return "", err
	}
	if task.Status != VideoStatusCompleted {
		return "", ErrVideoContentUnavailable
	}
	if task.ExpiresAt != nil && time.Now().UTC().After(*task.ExpiresAt) {
		_ = s.repo.MarkTaskExpired(ctx, task.PublicID)
		return "", ErrVideoContentExpired
	}
	for _, candidate := range []string{derefString(task.LocalContentURL), derefString(task.ContentURL), derefString(task.UpstreamContentURL)} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, err := normalizeVideoOutboundURL(candidate, s.cfg); err != nil {
			continue
		}
		return candidate, nil
	}
	return "", ErrVideoContentUnavailable
}

func (s *VideoService) ContentResponse(ctx context.Context, publicID string, apiKeyID int64) (*http.Response, error) {
	location, err := s.ContentLocation(ctx, publicID, apiKeyID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid content url", ErrVideoContentUnavailable)
	}
	resp, err := s.videoContentHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrVideoContentUnavailable, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("%w: upstream content returned %d", ErrVideoContentUnavailable, resp.StatusCode)
	}
	return resp, nil
}

func (s *VideoService) applyExpiredState(ctx context.Context, task *VideoGenerationTask) {
	if task == nil || task.Status != VideoStatusCompleted || task.ExpiresAt == nil {
		return
	}
	if time.Now().UTC().After(*task.ExpiresAt) {
		_ = s.repo.MarkTaskExpired(ctx, task.PublicID)
		s.releaseVideoSlots(task)
		task.Status = VideoStatusExpired
	}
}

func (s *VideoService) ProcessDueTasks(ctx context.Context, limit int) (int, error) {
	if s.repo == nil {
		return 0, nil
	}
	tasks, err := s.repo.ClaimDueTasks(ctx, limit, 2*time.Minute)
	if err != nil {
		return 0, err
	}
	for i := range tasks {
		if err := s.processTask(ctx, &tasks[i]); err != nil {
			if IsRetryableVideoError(err) {
				_ = s.scheduleRetry(ctx, &tasks[i], "upstream_retryable", err.Error())
				continue
			}
			_ = s.repo.MarkTaskFailed(ctx, tasks[i].PublicID, "worker_error", err.Error())
			s.releaseVideoSlots(&tasks[i])
		}
	}
	return len(tasks), nil
}

func (s *VideoService) processTask(ctx context.Context, task *VideoGenerationTask) error {
	if task == nil {
		return nil
	}
	switch task.Status {
	case VideoStatusQueued:
		return s.submitTask(ctx, task)
	case VideoStatusInProgress:
		return s.pollTask(ctx, task)
	default:
		return nil
	}
}

func (s *VideoService) submitTask(ctx context.Context, task *VideoGenerationTask) error {
	if s.upstreamClient == nil {
		return fmt.Errorf("video upstream client is not configured")
	}
	account, err := s.accountRepo.GetByID(ctx, task.AccountID)
	if err != nil {
		return err
	}
	template := taskTemplate(task)
	if template == nil {
		return ErrVideoTemplateNotFound
	}
	result, err := s.upstreamClient.Create(ctx, account, template, task.UpstreamRequestPayload)
	if err != nil {
		return err
	}
	next := time.Now().UTC().Add(videoPollInterval(template, task.PollCount))
	return s.repo.MarkTaskSubmitted(ctx, task.PublicID, result.TaskID, task.UpstreamRequestPayload, result.Response, next)
}

func (s *VideoService) pollTask(ctx context.Context, task *VideoGenerationTask) error {
	if s.upstreamClient == nil {
		return fmt.Errorf("video upstream client is not configured")
	}
	if task.UpstreamTaskID == nil || strings.TrimSpace(*task.UpstreamTaskID) == "" {
		return s.repo.MarkTaskFailed(ctx, task.PublicID, "missing_upstream_task_id", "missing upstream task id")
	}
	account, err := s.accountRepo.GetByID(ctx, task.AccountID)
	if err != nil {
		return err
	}
	template := taskTemplate(task)
	if template == nil {
		return ErrVideoTemplateNotFound
	}
	result, err := s.upstreamClient.Query(ctx, account, template, *task.UpstreamTaskID)
	if err != nil {
		return err
	}
	switch result.Status {
	case VideoStatusCompleted:
		if result.ContentURL == nil || strings.TrimSpace(*result.ContentURL) == "" {
			return s.repo.MarkTaskFailed(ctx, task.PublicID, "missing_content_url", "upstream completed without content url")
		}
		billableSeconds := task.RequestedSeconds
		if result.Seconds != nil && *result.Seconds > 0 {
			billableSeconds = result.Seconds
		}
		actual := calculateVideoActualCost(task, billableSeconds)
		expiresAt := time.Now().UTC().Add(s.contentRetention())
		if err := s.repo.MarkTaskCompleted(ctx, task.PublicID, 100, result.ContentURL, result.Response, billableSeconds, actual, expiresAt); err != nil {
			return err
		}
		s.releaseVideoSlots(task)
		if err := s.applyVideoFinalDelta(ctx, task, actual); err != nil {
			return err
		}
		return s.recordCompletedUsage(ctx, task, billableSeconds, actual)
	case VideoStatusFailed:
		code := result.ErrorCode
		if code == "" {
			code = "upstream_failed"
		}
		msg := result.ErrorMsg
		if msg == "" {
			msg = "upstream video generation failed"
		}
		err := s.repo.MarkTaskFailed(ctx, task.PublicID, code, msg)
		if err == nil {
			s.releaseVideoSlots(task)
		}
		return err
	case VideoStatusCancelled, VideoStatusExpired:
		_, err := s.repo.MarkTaskCancelled(ctx, task.PublicID)
		if err == nil {
			s.releaseVideoSlots(task)
		}
		return err
	default:
		if maxAttempts := videoMaxAttempts(template); maxAttempts > 0 && task.PollCount+1 >= maxAttempts {
			err := s.repo.MarkTaskFailed(ctx, task.PublicID, "poll_attempts_exceeded", "video task exceeded maximum poll attempts")
			if err == nil {
				s.releaseVideoSlots(task)
			}
			return err
		}
		next := time.Now().UTC().Add(videoPollInterval(template, task.PollCount+1))
		return s.repo.MarkTaskPollResult(ctx, task.PublicID, VideoStatusInProgress, clampProgress(result.Progress), result.Response, result.ContentURL, result.Response, &next)
	}
}

func (s *VideoService) scheduleRetry(ctx context.Context, task *VideoGenerationTask, code, message string) error {
	if task == nil {
		return nil
	}
	template := taskTemplate(task)
	if template == nil {
		return ErrVideoTemplateNotFound
	}
	if maxAttempts := videoMaxAttempts(template); maxAttempts > 0 && task.PollCount+1 >= maxAttempts {
		err := s.repo.MarkTaskFailed(ctx, task.PublicID, "retry_attempts_exceeded", "video task exceeded maximum retry attempts")
		if err == nil {
			s.releaseVideoSlots(task)
		}
		return err
	}
	next := time.Now().UTC().Add(videoPollInterval(template, task.PollCount+1))
	return s.repo.ScheduleTaskRetry(ctx, task.PublicID, next, code, message)
}

func (s *VideoService) recordCompletedUsage(ctx context.Context, task *VideoGenerationTask, billableSeconds *int, actualCost float64) error {
	if s.usageLogRepo != nil {
		units := calculateVideoBillingUnits(task, billableSeconds)
		size := ""
		if v, ok := task.RequestPayload["size"].(string); ok {
			size = v
		}
		billingMode := task.BillingMode
		_, _ = s.usageLogRepo.Create(ctx, &UsageLog{
			UserID:            task.UserID,
			APIKeyID:          task.APIKeyID,
			AccountID:         task.AccountID,
			RequestID:         "videolog:" + task.PublicID,
			Model:             task.UpstreamModel,
			RequestedModel:    task.RequestedModel,
			ChannelID:         task.ChannelID,
			GroupID:           task.GroupID,
			TotalCost:         actualCost,
			ActualCost:        actualCost,
			RateMultiplier:    1,
			BillingType:       BillingTypeBalance,
			RequestType:       RequestTypeSync,
			BillingMode:       &billingMode,
			InboundEndpoint:   stringPtr("/v1/videos"),
			VideoTaskID:       &task.PublicID,
			VideoSeconds:      billableSeconds,
			VideoSize:         nilIfEmpty(size),
			VideoBillingUnits: &units,
		})
	}
	return nil
}

func (s *VideoService) applyVideoFinalDelta(ctx context.Context, task *VideoGenerationTask, actualCost float64) error {
	if s.usageBilling == nil || task == nil {
		return nil
	}
	delta := actualCost - task.ReservedCost
	if delta <= 0 {
		return nil
	}
	_, err := s.usageBilling.Apply(ctx, &UsageBillingCommand{
		RequestID:        "videofin:" + task.PublicID,
		UserID:           task.UserID,
		APIKeyID:         task.APIKeyID,
		AccountID:        task.AccountID,
		BalanceCost:      delta,
		APIKeyQuotaCost:  delta,
		AccountQuotaCost: delta,
		RequestFingerprint: usageBillingFingerprint(map[string]any{
			"public_id": task.PublicID,
			"model":     task.RequestedModel,
			"stage":     "final_delta",
			"delta":     delta,
		}),
	})
	return err
}

func (s *VideoService) ListModels(ctx context.Context, groupID *int64) (*VideoModelListResponse, error) {
	models, err := s.repo.ListActiveModels(ctx)
	if err != nil {
		return nil, err
	}
	resp := &VideoModelListResponse{Object: "list", Data: make([]VideoModelObject, 0, len(models))}
	for i := range models {
		m := &models[i]
		obj := VideoModelObject{
			ID:             m.PublicModel,
			Object:         "model",
			DisplayName:    derefString(m.DisplayName),
			Status:         m.Status,
			Supports:       stringSliceFromAny(m.Capabilities["supports"]),
			Seconds:        intSliceFromAny(m.SupportedOptions["seconds"]),
			Sizes:          stringSliceFromAny(m.SupportedOptions["sizes"]),
			Limits:         m.Limits,
			ExtraBodyAllow: append([]string(nil), m.ExtraBodyAllow...),
		}
		if p, err := s.resolveVideoPricing(ctx, groupID, m.PublicModel, firstOrDefault(obj.Seconds, intFromAny(m.Defaults["seconds"]))); err == nil {
			obj.Billing = &VideoBilling{
				Mode:      string(p.Mode),
				UnitPrice: p.UnitPrice,
				Currency:  "USD",
			}
			if p.UnitSeconds != nil {
				obj.Billing.UnitSeconds = *p.UnitSeconds
			}
		}
		resp.Data = append(resp.Data, obj)
	}
	return resp, nil
}

func (s *VideoService) AdminListTemplates(ctx context.Context) ([]VideoCallTemplate, error) {
	return s.repo.ListTemplates(ctx)
}

func (s *VideoService) AdminCreateTemplate(ctx context.Context, template *VideoCallTemplate) (*VideoCallTemplate, error) {
	if err := validateVideoTemplate(template); err != nil {
		return nil, err
	}
	if err := s.repo.CreateTemplate(ctx, template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *VideoService) AdminUpdateTemplate(ctx context.Context, id int64, template *VideoCallTemplate) (*VideoCallTemplate, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid template id", ErrVideoInvalidRequest)
	}
	template.ID = id
	if err := validateVideoTemplate(template); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateTemplate(ctx, template); err != nil {
		return nil, err
	}
	return s.repo.GetTemplateByID(ctx, id)
}

func (s *VideoService) AdminDeleteTemplate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid template id", ErrVideoInvalidRequest)
	}
	models, err := s.repo.ListModels(ctx, true)
	if err != nil {
		return err
	}
	var refs []string
	for i := range models {
		if models[i].TemplateID == id {
			refs = append(refs, models[i].PublicModel)
		}
	}
	if len(refs) > 0 {
		return fmt.Errorf("%w: template is used by video models: %s", ErrVideoInvalidRequest, strings.Join(refs, ", "))
	}
	return s.repo.DeleteTemplate(ctx, id)
}

func (s *VideoService) AdminTestTemplateCreate(ctx context.Context, accountID int64, templateID int64, body map[string]any) (*VideoUpstreamCreateResult, error) {
	if accountID <= 0 || templateID <= 0 {
		return nil, fmt.Errorf("%w: account_id and template_id are required", ErrVideoInvalidRequest)
	}
	if body == nil {
		body = map[string]any{}
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	template, err := s.repo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if s.upstreamClient == nil {
		return nil, fmt.Errorf("video upstream client is not configured")
	}
	return s.upstreamClient.Create(ctx, account, template, body)
}

func (s *VideoService) AdminTestTemplateQuery(ctx context.Context, accountID int64, templateID int64, upstreamTaskID string) (*VideoUpstreamQueryResult, error) {
	if accountID <= 0 || templateID <= 0 || strings.TrimSpace(upstreamTaskID) == "" {
		return nil, fmt.Errorf("%w: account_id, template_id and upstream_task_id are required", ErrVideoInvalidRequest)
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	template, err := s.repo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if s.upstreamClient == nil {
		return nil, fmt.Errorf("video upstream client is not configured")
	}
	return s.upstreamClient.Query(ctx, account, template, strings.TrimSpace(upstreamTaskID))
}

func (s *VideoService) AdminListModels(ctx context.Context) ([]VideoModel, error) {
	return s.repo.ListModels(ctx, true)
}

func (s *VideoService) AdminRequestShapes(ctx context.Context) ([]string, error) {
	return RegisteredVideoRequestShapes(), nil
}

// AdminRecognizeTemplate uses the chosen video account's OpenAI-compatible chat
// endpoint to parse pasted upstream API docs into a draft template. The result
// is not persisted; the admin reviews and saves it via the normal create flow.
func (s *VideoService) AdminRecognizeTemplate(ctx context.Context, accountID int64, model, doc string) (*VideoCallTemplate, error) {
	if accountID <= 0 {
		return nil, fmt.Errorf("%w: account_id is required", ErrVideoInvalidRequest)
	}
	if strings.TrimSpace(doc) == "" {
		return nil, fmt.Errorf("%w: document is required", ErrVideoInvalidRequest)
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("%w: model is required", ErrVideoInvalidRequest)
	}
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil || account.Platform != PlatformVideo {
		return nil, fmt.Errorf("%w: account must be a video platform account", ErrVideoInvalidRequest)
	}
	if account.Type != AccountTypeAPIKey {
		return nil, fmt.Errorf("%w: account must be a video API key account", ErrVideoInvalidRequest)
	}
	if account.Status != StatusActive {
		return nil, fmt.Errorf("%w: account must be active", ErrVideoInvalidRequest)
	}
	if s.upstreamClient == nil {
		return nil, fmt.Errorf("video upstream client is not configured")
	}
	template, err := s.upstreamClient.RecognizeTemplate(ctx, account, strings.TrimSpace(model), doc)
	if err != nil {
		return nil, err
	}
	// Fail closed: the recognized template must satisfy the same validation as a
	// hand-written one before we hand it back to the admin.
	if err := validateVideoTemplate(template); err != nil {
		return nil, err
	}
	return template, nil
}

func (s *VideoService) AdminCreateModel(ctx context.Context, model *VideoModel) (*VideoModel, error) {
	if err := validateVideoModel(model); err != nil {
		return nil, err
	}
	if err := s.repo.CreateModel(ctx, model); err != nil {
		return nil, err
	}
	return s.repo.GetModelByID(ctx, model.ID)
}

func (s *VideoService) AdminUpdateModel(ctx context.Context, id int64, model *VideoModel) (*VideoModel, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid model id", ErrVideoInvalidRequest)
	}
	model.ID = id
	if err := validateVideoModel(model); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateModel(ctx, model); err != nil {
		return nil, err
	}
	return s.repo.GetModelByID(ctx, id)
}

func (s *VideoService) AdminDeleteModel(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid model id", ErrVideoInvalidRequest)
	}
	return s.repo.DeleteModel(ctx, id)
}

func (s *VideoService) AdminListTasks(ctx context.Context, filter VideoTaskFilter) ([]VideoGenerationTask, int64, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	items, total, err := s.repo.ListTasks(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		s.applyExpiredState(ctx, &items[i])
	}
	return items, total, nil
}

func (s *VideoService) AdminGetTask(ctx context.Context, publicID string) (*VideoGenerationTask, error) {
	task, err := s.repo.GetTaskByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	s.applyExpiredState(ctx, task)
	return task, nil
}

func (s *VideoService) AdminRequeueTask(ctx context.Context, publicID string) (*VideoGenerationTask, error) {
	task, err := s.repo.GetTaskByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	if task.Status != VideoStatusQueued && task.Status != VideoStatusInProgress && task.Status != VideoStatusFailed {
		return nil, fmt.Errorf("%w: only queued, in_progress or failed tasks can be requeued", ErrVideoInvalidRequest)
	}
	if err := s.repo.RequeueTask(ctx, publicID, time.Now().UTC()); err != nil {
		return nil, err
	}
	return s.repo.GetTaskByPublicID(ctx, publicID)
}

func (s *VideoService) AdminFailTask(ctx context.Context, publicID string, code, message string) (*VideoGenerationTask, error) {
	if strings.TrimSpace(code) == "" {
		code = "admin_forced_failed"
	}
	if strings.TrimSpace(message) == "" {
		message = "video task was forced failed by admin"
	}
	if err := s.repo.MarkTaskFailed(ctx, publicID, code, message); err != nil {
		return nil, err
	}
	task, err := s.repo.GetTaskByPublicID(ctx, publicID)
	if err == nil {
		s.releaseVideoSlots(task)
	}
	return task, err
}

func (s *VideoService) acquireVideoSlots(ctx context.Context, task *VideoGenerationTask, input VideoCreateInput, account *Account) (*AcquireResult, error) {
	if s.concurrency == nil || task == nil || input.User == nil || input.APIKey == nil || account == nil {
		return &AcquireResult{Acquired: true, ReleaseFunc: func() {}}, nil
	}
	userMax := input.User.Concurrency
	keyMax := input.User.Concurrency
	accountMax := account.Concurrency
	return s.concurrency.AcquireVideoSlots(ctx, task.PublicID, input.User.ID, input.APIKey.ID, account.ID, userMax, keyMax, accountMax)
}

func (s *VideoService) releaseVideoSlots(task *VideoGenerationTask) {
	if s == nil || s.concurrency == nil || task == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.concurrency.ReleaseVideoSlots(ctx, task.PublicID, task.UserID, task.APIKeyID, task.AccountID)
}

func (s *VideoService) reserveTaskCost(ctx context.Context, task *VideoGenerationTask, input VideoCreateInput) error {
	if s.usageBilling == nil || task.EstimatedCost <= 0 {
		return nil
	}
	_, err := s.usageBilling.Apply(ctx, &UsageBillingCommand{
		RequestID:        "videores:" + task.PublicID,
		UserID:           input.User.ID,
		APIKeyID:         input.APIKey.ID,
		AccountID:        task.AccountID,
		BalanceCost:      task.EstimatedCost,
		APIKeyQuotaCost:  task.EstimatedCost,
		AccountQuotaCost: task.EstimatedCost,
		RequestFingerprint: usageBillingFingerprint(map[string]any{
			"public_id": task.PublicID,
			"model":     task.RequestedModel,
			"stage":     "reserve",
		}),
	})
	return err
}

func (s *VideoService) selectVideoAccount(ctx context.Context, groupID *int64, group *Group, model string) (*Account, error) {
	var accounts []Account
	var err error
	platforms := videoCandidatePlatforms(group)
	if groupID != nil {
		accounts, err = s.accountRepo.ListSchedulableByGroupIDAndPlatforms(ctx, *groupID, platforms)
	} else {
		accounts, err = s.accountRepo.ListSchedulableUngroupedByPlatforms(ctx, platforms)
	}
	if err != nil {
		return nil, err
	}
	accounts = filterVideoCapableAccounts(accounts, model)
	if len(accounts) == 0 {
		return nil, ErrNoAvailableAccounts
	}
	if routed := selectRoutedVideoAccount(accounts, group, model); routed != nil {
		full, err := s.accountRepo.GetByID(ctx, routed.ID)
		if err != nil {
			return nil, err
		}
		return full, nil
	}
	best := accounts[0]
	for _, acc := range accounts[1:] {
		if acc.Priority < best.Priority || (acc.Priority == best.Priority && acc.ID < best.ID) {
			best = acc
		}
	}
	full, err := s.accountRepo.GetByID(ctx, best.ID)
	if err != nil {
		return nil, err
	}
	return full, nil
}

func videoCandidatePlatforms(group *Group) []string {
	platforms := []string{PlatformVideo}
	if group == nil {
		return platforms
	}
	add := func(platform string) {
		platform = strings.ToLower(strings.TrimSpace(platform))
		if platform == "" {
			return
		}
		for _, existing := range platforms {
			if existing == platform {
				return
			}
		}
		platforms = append(platforms, platform)
	}
	for _, scope := range group.SupportedModelScopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if strings.HasPrefix(scope, "video:") {
			add(strings.TrimPrefix(scope, "video:"))
		}
	}
	return platforms
}

func filterVideoCapableAccounts(accounts []Account, model string) []Account {
	if len(accounts) == 0 {
		return nil
	}
	filtered := make([]Account, 0, len(accounts))
	for _, account := range accounts {
		if !account.SupportsCapability(AccountCapabilityVideos) {
			continue
		}
		if !account.IsModelSupported(model) {
			continue
		}
		filtered = append(filtered, account)
	}
	return filtered
}

func selectRoutedVideoAccount(accounts []Account, group *Group, model string) *Account {
	if len(accounts) == 0 || group == nil {
		return nil
	}
	routedIDs := group.GetRoutingAccountIDs(model)
	if len(routedIDs) == 0 {
		return nil
	}
	byID := make(map[int64]Account, len(accounts))
	for _, acc := range accounts {
		byID[acc.ID] = acc
	}
	for _, id := range routedIDs {
		if acc, ok := byID[id]; ok {
			return &acc
		}
	}
	return nil
}

func (s *VideoService) contentRetention() time.Duration {
	if s != nil && s.cfg != nil && s.cfg.Gateway.VideoContentRetentionHours > 0 {
		return time.Duration(s.cfg.Gateway.VideoContentRetentionHours) * time.Hour
	}
	return 24 * time.Hour
}

func (s *VideoService) videoContentHTTPClient() *http.Client {
	var cfg *config.Config
	if s != nil {
		cfg = s.cfg
	}
	return newVideoHTTPClient(cfg, 10*time.Minute)
}

func newVideoHTTPClient(cfg *config.Config, timeout time.Duration) *http.Client {
	if cfg == nil || !cfg.Security.URLAllowlist.Enabled || cfg.Security.URLAllowlist.AllowPrivateHosts {
		return &http.Client{Timeout: timeout}
	}
	transport, _ := videoContentHTTPClient.Transport.(*http.Transport)
	if transport == nil {
		return &http.Client{Timeout: timeout}
	}
	return &http.Client{Timeout: timeout, Transport: transport.Clone()}
}

type videoPricing struct {
	Mode          BillingMode
	UnitPrice     float64
	UnitSeconds   *float64
	EstimatedCost float64
}

func (s *VideoService) resolveVideoPricing(ctx context.Context, groupID *int64, model string, seconds int) (*videoPricing, error) {
	if s.pricingResolver == nil {
		return nil, ErrVideoPricingUnavailable
	}
	resolved := s.pricingResolver.Resolve(ctx, PricingInput{Model: model, GroupID: groupID})
	mode := resolved.Mode
	switch mode {
	case BillingModePerRequest, BillingModeImage:
		unit := resolved.DefaultPerRequestPrice
		return &videoPricing{Mode: BillingModePerRequest, UnitPrice: unit, EstimatedCost: unit}, nil
	case BillingModeSecond:
		unit := resolved.DefaultPerRequestPrice
		return &videoPricing{Mode: BillingModeSecond, UnitPrice: unit, EstimatedCost: unit * float64(seconds)}, nil
	case BillingModeSegment:
		unit := resolved.DefaultPerRequestPrice
		unitSeconds := resolved.UnitSeconds
		if unitSeconds <= 0 {
			return nil, ErrVideoPricingUnavailable
		}
		segments := math.Ceil(float64(seconds) / unitSeconds)
		return &videoPricing{Mode: BillingModeSegment, UnitPrice: unit, UnitSeconds: &unitSeconds, EstimatedCost: unit * segments}, nil
	default:
		return nil, ErrVideoPricingUnavailable
	}
}

func TaskToVideoObject(task *VideoGenerationTask) VideoObject {
	obj := VideoObject{
		ID:        task.PublicID,
		Object:    VideoObjectType,
		Model:     task.RequestedModel,
		Status:    task.Status,
		Progress:  task.Progress,
		CreatedAt: task.CreatedAt.Unix(),
		Error:     nil,
	}
	if task.CompletedAt != nil {
		value := task.CompletedAt.Unix()
		obj.CompletedAt = &value
	}
	if task.ExpiresAt != nil {
		value := task.ExpiresAt.Unix()
		obj.ExpiresAt = &value
	}
	if task.RequestedSeconds != nil {
		obj.Seconds = strconv.Itoa(*task.RequestedSeconds)
	}
	if size, ok := task.RequestPayload["size"].(string); ok {
		obj.Size = size
	}
	if task.ErrorCode != nil || task.ErrorMessage != nil {
		obj.Error = &VideoObjectError{Code: derefString(task.ErrorCode), Message: derefString(task.ErrorMessage)}
	}
	return obj
}

func ParseVideoCreateRequest(body []byte) (VideoCreateRequest, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return VideoCreateRequest{}, err
	}
	var req VideoCreateRequest
	_ = json.Unmarshal(body, &req)
	if req.ExtraBody == nil {
		req.ExtraBody = map[string]any{}
	}
	req.Raw = raw
	return req, nil
}

type ShapeBuilder func(req *VideoCreateRequest, model *VideoModel, upstreamModel string) (map[string]any, error)

var videoShapeBuilders = map[string]ShapeBuilder{
	"videos":       buildVideosShape,
	"seedance":     buildSeedanceShape,
	"grok_imagine": buildGrokImagineShape,
}

func RegisteredVideoRequestShapes() []string {
	shapes := make([]string, 0, len(videoShapeBuilders))
	for shape := range videoShapeBuilders {
		shapes = append(shapes, shape)
	}
	sort.Strings(shapes)
	return shapes
}

// BuildVideoUpstreamRequest builds the upstream create-task body. upstreamModel
// is the model name sent to the upstream — resolved from the account-level model
// mapping (public model → upstream model), not from video_models.
func BuildVideoUpstreamRequest(req *VideoCreateRequest, model *VideoModel, upstreamModel string) (map[string]any, error) {
	builder, ok := videoShapeBuilders[model.RequestShape]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported request_shape %s", ErrVideoInvalidRequest, model.RequestShape)
	}
	return builder(req, model, upstreamModel)
}

func buildVideosShape(req *VideoCreateRequest, model *VideoModel, upstreamModel string) (map[string]any, error) {
	seconds, err := resolveVideoSeconds(*req, model)
	if err != nil {
		return nil, err
	}
	_, resolution, ratio, err := resolveVideoSize(*req, model)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"model":      upstreamModel,
		"prompt":     req.Prompt,
		"duration":   seconds,
		"ratio":      ratio,
		"resolution": strings.ToLower(resolution),
	}
	mergeAllowedExtra(body, req.ExtraBody, model.ExtraBodyAllow)
	return body, nil
}

func buildSeedanceShape(req *VideoCreateRequest, model *VideoModel, upstreamModel string) (map[string]any, error) {
	seconds, err := resolveVideoSeconds(*req, model)
	if err != nil {
		return nil, err
	}
	_, resolution, ratio, err := resolveVideoSize(*req, model)
	if err != nil {
		return nil, err
	}
	metadata := map[string]any{
		"resolution": resolution,
		"ratio":      ratio,
	}
	body := map[string]any{
		"model":    upstreamModel,
		"prompt":   req.Prompt,
		"duration": seconds,
		"metadata": metadata,
	}
	if req.InputReference != "" {
		body["input_reference"] = req.InputReference
	}
	mergeAllowedExtra(metadata, req.ExtraBody, model.ExtraBodyAllow)
	return body, nil
}

func buildGrokImagineShape(req *VideoCreateRequest, model *VideoModel, upstreamModel string) (map[string]any, error) {
	if strings.TrimSpace(req.InputReference) == "" {
		return nil, fmt.Errorf("%w: input_reference is required for %s", ErrVideoInvalidRequest, model.PublicModel)
	}
	seconds, err := resolveVideoSeconds(*req, model)
	if err != nil {
		return nil, err
	}
	size, _, _, err := resolveVideoSize(*req, model)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"model":  upstreamModel,
		"prompt": req.Prompt,
		// OpenAI-style video endpoints (e.g. grok-imagine via aggregators) expect
		// the standard `input_reference` field; keep `image_url` for grok variants
		// that use that name instead.
		"input_reference": req.InputReference,
		"image_url":       req.InputReference,
		"seconds":         strconv.Itoa(seconds),
		"size":            size,
	}
	mergeAllowedExtra(body, req.ExtraBody, model.ExtraBodyAllow)
	return body, nil
}

func resolveVideoSeconds(req VideoCreateRequest, model *VideoModel) (int, error) {
	raw := strings.TrimSpace(req.Seconds)
	if raw == "" {
		raw = strings.TrimSpace(videoStringFromAny(model.Defaults["seconds"]))
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%w: seconds must be a positive integer string", ErrVideoInvalidRequest)
	}
	if allowed := intSliceFromAny(model.SupportedOptions["seconds"]); len(allowed) > 0 && !containsInt(allowed, seconds) {
		return 0, fmt.Errorf("%w: seconds is not supported", ErrVideoInvalidRequest)
	}
	return seconds, nil
}

func resolveVideoSize(req VideoCreateRequest, model *VideoModel) (size, resolution, ratio string, err error) {
	size = strings.TrimSpace(req.Size)
	if size == "" {
		size = strings.TrimSpace(videoStringFromAny(model.Defaults["size"]))
	}
	if size == "" {
		return "", "", "", fmt.Errorf("%w: size is required", ErrVideoInvalidRequest)
	}
	if allowed := stringSliceFromAny(model.SupportedOptions["sizes"]); len(allowed) > 0 && !containsStringFold(allowed, size) {
		return "", "", "", fmt.Errorf("%w: size is not supported", ErrVideoInvalidRequest)
	}
	parts := strings.Split(strings.ToLower(size), "x")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("%w: size must be WIDTHxHEIGHT", ErrVideoInvalidRequest)
	}
	w, wErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, hErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if wErr != nil || hErr != nil || w <= 0 || h <= 0 {
		return "", "", "", fmt.Errorf("%w: size must be WIDTHxHEIGHT", ErrVideoInvalidRequest)
	}
	ratio = reduceRatio(w, h)
	maxDim := w
	if h > maxDim {
		maxDim = h
	}
	switch {
	case maxDim <= 854:
		resolution = "480P"
	case maxDim <= 1280:
		resolution = "720P"
	case maxDim <= 1920:
		resolution = "1080P"
	default:
		return "", "", "", fmt.Errorf("%w: size exceeds supported resolution", ErrVideoInvalidRequest)
	}
	return size, resolution, ratio, nil
}

func mergeAllowedExtra(dst map[string]any, extra map[string]any, allowed []string) {
	if len(extra) == 0 || len(allowed) == 0 {
		return
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key, value := range extra {
		if key == "model" {
			continue
		}
		if _, ok := allowedSet[key]; ok {
			dst[key] = value
		}
	}
}

func (s *VideoService) ValidateVideoURLs(req VideoCreateRequest) error {
	for _, raw := range []string{req.InputReference} {
		if raw == "" {
			continue
		}
		if _, err := normalizeVideoOutboundURL(raw, s.cfg); err != nil {
			return fmt.Errorf("%w: %s", ErrVideoInvalidRequest, err.Error())
		}
	}
	var validateExtra func(any) error
	validateExtra = func(v any) error {
		switch typed := v.(type) {
		case string:
			if !looksLikeHTTPURL(typed) {
				return nil
			}
			if _, err := normalizeVideoOutboundURL(typed, s.cfg); err != nil {
				return fmt.Errorf("%w: %s", ErrVideoInvalidRequest, err.Error())
			}
		case []any:
			for _, item := range typed {
				if err := validateExtra(item); err != nil {
					return err
				}
			}
		case map[string]any:
			for _, item := range typed {
				if err := validateExtra(item); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return validateExtra(req.ExtraBody)
}

func ValidateVideoURLs(req VideoCreateRequest) error {
	return (&VideoService{}).ValidateVideoURLs(req)
}

func normalizeVideoOutboundURL(raw string, cfg *config.Config) (string, error) {
	if cfg == nil || !cfg.Security.URLAllowlist.Enabled {
		allowHTTP := true
		if cfg != nil {
			allowHTTP = cfg.Security.URLAllowlist.AllowInsecureHTTP
		}
		return urlvalidator.ValidateURLFormat(raw, allowHTTP)
	}
	return urlvalidator.ValidateHTTPURL(raw, cfg.Security.URLAllowlist.AllowInsecureHTTP, urlvalidator.ValidationOptions{
		AllowedHosts:     cfg.Security.URLAllowlist.UpstreamHosts,
		RequireAllowlist: true,
		AllowPrivate:     cfg.Security.URLAllowlist.AllowPrivateHosts,
	})
}

func looksLikeHTTPURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return (scheme == "http" || scheme == "https") && u.Host != ""
}

func reserveVideoSeconds(model *VideoModel, requested int) int {
	if model == nil || requested <= 0 {
		return requested
	}
	maxSeconds := intFromAny(model.Limits["max_seconds"])
	if maxSeconds <= 0 {
		for _, seconds := range intSliceFromAny(model.SupportedOptions["seconds"]) {
			if seconds > maxSeconds {
				maxSeconds = seconds
			}
		}
	}
	if maxSeconds > requested {
		return maxSeconds
	}
	return requested
}

func validateVideoTemplate(template *VideoCallTemplate) error {
	if template == nil {
		return fmt.Errorf("%w: template is required", ErrVideoInvalidRequest)
	}
	template.Name = strings.TrimSpace(template.Name)
	template.CreateMethod = strings.ToUpper(strings.TrimSpace(defaultString(template.CreateMethod, httpMethodPost)))
	template.QueryMethod = strings.ToUpper(strings.TrimSpace(defaultString(template.QueryMethod, httpMethodGet)))
	template.CreatePath = strings.TrimSpace(template.CreatePath)
	template.QueryPath = strings.TrimSpace(template.QueryPath)
	template.Status = strings.TrimSpace(defaultString(template.Status, "active"))
	if template.Name == "" || template.CreatePath == "" || template.QueryPath == "" {
		return fmt.Errorf("%w: template name/create_path/query_path are required", ErrVideoInvalidRequest)
	}
	if template.StatusMapping == nil {
		template.StatusMapping = map[string]string{}
	}
	normalized := make(map[string]string, len(template.StatusMapping))
	for k, v := range template.StatusMapping {
		normalized[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	template.StatusMapping = normalized
	if template.ResultMapping == nil {
		template.ResultMapping = map[string]string{}
	}
	if template.ErrorMapping == nil {
		template.ErrorMapping = map[string]string{}
	}
	if template.PollConfig == nil {
		template.PollConfig = map[string]any{}
	}
	if template.TimeoutConfig == nil {
		template.TimeoutConfig = map[string]any{}
	}
	return nil
}

func validateVideoModel(model *VideoModel) error {
	if model == nil {
		return fmt.Errorf("%w: model is required", ErrVideoInvalidRequest)
	}
	model.PublicModel = strings.TrimSpace(model.PublicModel)
	model.UpstreamModelID = strings.TrimSpace(model.UpstreamModelID)
	model.RequestShape = strings.TrimSpace(model.RequestShape)
	model.Status = strings.TrimSpace(defaultString(model.Status, "active"))
	// upstream_model_id is no longer required here: the public→upstream model
	// name mapping is configured on the video Account (credentials.model_mapping),
	// matching how other platforms resolve upstream model names.
	if model.PublicModel == "" || model.TemplateID <= 0 || model.RequestShape == "" {
		return fmt.Errorf("%w: public_model/template_id/request_shape are required", ErrVideoInvalidRequest)
	}
	if _, ok := videoShapeBuilders[model.RequestShape]; !ok {
		return fmt.Errorf("%w: unsupported request_shape %s", ErrVideoInvalidRequest, model.RequestShape)
	}
	if model.Capabilities == nil {
		model.Capabilities = map[string]any{}
	}
	if model.Defaults == nil {
		model.Defaults = map[string]any{}
	}
	if model.Limits == nil {
		model.Limits = map[string]any{}
	}
	if model.SupportedOptions == nil {
		model.SupportedOptions = map[string]any{}
	}
	if model.ExtraBodyAllow == nil {
		model.ExtraBodyAllow = []string{}
	}
	return nil
}

const (
	httpMethodGet  = "GET"
	httpMethodPost = "POST"
)

func videoIdempotencyHash(apiKeyID int64, key string) string {
	sum := sha256.Sum256([]byte(strconv.FormatInt(apiKeyID, 10) + ":" + key))
	return hex.EncodeToString(sum[:])
}

func usageBillingFingerprint(v any) string {
	body, _ := json.Marshal(v)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func reduceRatio(w, h int) string {
	g := gcd(w, h)
	return fmt.Sprintf("%d:%d", w/g, h/g)
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func videoPtrTime(t time.Time) *time.Time { return &t }

func nilIfEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func firstOrDefault(values []int, fallback int) int {
	if fallback > 0 {
		return fallback
	}
	if len(values) > 0 {
		return values[0]
	}
	return 1
}

func taskTemplate(task *VideoGenerationTask) *VideoCallTemplate {
	if task == nil || task.VideoModel == nil {
		return nil
	}
	return task.VideoModel.Template
}

func videoPollInterval(template *VideoCallTemplate, pollCount int) time.Duration {
	seconds := intFromAny(template.PollConfig["interval_seconds"])
	if seconds <= 0 {
		seconds = 5
	}
	maxSeconds := intFromAny(template.PollConfig["backoff_max_seconds"])
	if maxSeconds <= 0 {
		maxSeconds = 30
	}
	backoff := seconds
	if pollCount > 0 {
		backoff = seconds << minInt(pollCount, 4)
	}
	if backoff > maxSeconds {
		backoff = maxSeconds
	}
	return time.Duration(backoff) * time.Second
}

func videoMaxAttempts(template *VideoCallTemplate) int {
	if template == nil {
		return 0
	}
	return intFromAny(template.PollConfig["max_attempts"])
}

func calculateVideoActualCost(task *VideoGenerationTask, billableSeconds *int) float64 {
	if task == nil {
		return 0
	}
	seconds := 1
	if billableSeconds != nil && *billableSeconds > 0 {
		seconds = *billableSeconds
	} else if task.RequestedSeconds != nil && *task.RequestedSeconds > 0 {
		seconds = *task.RequestedSeconds
	}
	switch BillingMode(task.BillingMode) {
	case BillingModeSecond:
		return task.UnitPrice * float64(seconds)
	case BillingModeSegment:
		unitSeconds := 1.0
		if task.UnitSeconds != nil && *task.UnitSeconds > 0 {
			unitSeconds = *task.UnitSeconds
		}
		return task.UnitPrice * math.Ceil(float64(seconds)/unitSeconds)
	default:
		return task.UnitPrice
	}
}

func calculateVideoBillingUnits(task *VideoGenerationTask, billableSeconds *int) int {
	if task == nil {
		return 1
	}
	seconds := 1
	if billableSeconds != nil && *billableSeconds > 0 {
		seconds = *billableSeconds
	}
	if BillingMode(task.BillingMode) == BillingModeSegment {
		unitSeconds := 1.0
		if task.UnitSeconds != nil && *task.UnitSeconds > 0 {
			unitSeconds = *task.UnitSeconds
		}
		return int(math.Ceil(float64(seconds) / unitSeconds))
	}
	if BillingMode(task.BillingMode) == BillingModeSecond {
		return seconds
	}
	return 1
}

func clampProgress(v int) int {
	if v < 0 {
		return 0
	}
	if v > 99 {
		return 99
	}
	return v
}

func stringPtr(v string) *string { return &v }

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsStringFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		return 0
	}
}

func videoStringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.Itoa(int(t))
	case int:
		return strconv.Itoa(t)
	default:
		return ""
	}
}

func stringSliceFromAny(v any) []string {
	items, ok := v.([]any)
	if !ok {
		if typed, ok := v.([]string); ok {
			return typed
		}
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func intSliceFromAny(v any) []int {
	items, ok := v.([]any)
	if !ok {
		if typed, ok := v.([]int); ok {
			return typed
		}
		return nil
	}
	out := make([]int, 0, len(items))
	for _, item := range items {
		switch t := item.(type) {
		case float64:
			out = append(out, int(t))
		case int:
			out = append(out, t)
		case string:
			if n, err := strconv.Atoi(t); err == nil {
				out = append(out, n)
			}
		}
	}
	return out
}
