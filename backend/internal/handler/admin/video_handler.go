package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoService *service.VideoService
}

func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{videoService: videoService}
}

type videoTemplateRequest struct {
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
	Status        string            `json:"status"`
}

type videoModelRequest struct {
	PublicModel      string         `json:"public_model"`
	DisplayName      *string        `json:"display_name"`
	TemplateID       int64          `json:"template_id"`
	UpstreamModelID  string         `json:"upstream_model_id"`
	RequestShape     string         `json:"request_shape"`
	Status           string         `json:"status"`
	Capabilities     map[string]any `json:"capabilities"`
	Defaults         map[string]any `json:"defaults"`
	Limits           map[string]any `json:"limits"`
	SupportedOptions map[string]any `json:"supported_options"`
	ExtraBodyAllow   []string       `json:"extra_body_allow"`
	SortOrder        int            `json:"sort_order"`
}

type videoTaskActionRequest struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type videoTemplateCreateTestRequest struct {
	AccountID  int64          `json:"account_id"`
	TemplateID int64          `json:"template_id"`
	Body       map[string]any `json:"body"`
}

type videoTemplateQueryTestRequest struct {
	AccountID      int64  `json:"account_id"`
	TemplateID     int64  `json:"template_id"`
	UpstreamTaskID string `json:"upstream_task_id"`
}

type videoTemplateRecognizeRequest struct {
	AccountID int64  `json:"account_id"`
	Model     string `json:"model"`
	Document  string `json:"document"`
}

type videoTemplateResponse struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	CreateMethod  string            `json:"create_method"`
	CreatePath    string            `json:"create_path"`
	QueryMethod   string            `json:"query_method"`
	QueryPath     string            `json:"query_path"`
	ContentMethod *string           `json:"content_method,omitempty"`
	ContentPath   *string           `json:"content_path,omitempty"`
	CancelMethod  *string           `json:"cancel_method,omitempty"`
	CancelPath    *string           `json:"cancel_path,omitempty"`
	StatusMapping map[string]string `json:"status_mapping"`
	ResultMapping map[string]string `json:"result_mapping"`
	ErrorMapping  map[string]string `json:"error_mapping"`
	PollConfig    map[string]any    `json:"poll_config"`
	TimeoutConfig map[string]any    `json:"timeout_config"`
	Status        string            `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type videoModelResponse struct {
	ID               int64                  `json:"id"`
	PublicModel      string                 `json:"public_model"`
	DisplayName      *string                `json:"display_name,omitempty"`
	TemplateID       int64                  `json:"template_id"`
	UpstreamModelID  string                 `json:"upstream_model_id"`
	RequestShape     string                 `json:"request_shape"`
	Status           string                 `json:"status"`
	Capabilities     map[string]any         `json:"capabilities"`
	Defaults         map[string]any         `json:"defaults"`
	Limits           map[string]any         `json:"limits"`
	SupportedOptions map[string]any         `json:"supported_options"`
	ExtraBodyAllow   []string               `json:"extra_body_allow"`
	SortOrder        int                    `json:"sort_order"`
	Template         *videoTemplateResponse `json:"template,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

type videoTaskResponse struct {
	ID                      int64               `json:"id"`
	PublicID                string              `json:"public_id"`
	UserID                  int64               `json:"user_id"`
	APIKeyID                int64               `json:"api_key_id"`
	GroupID                 *int64              `json:"group_id,omitempty"`
	AccountID               int64               `json:"account_id"`
	ChannelID               *int64              `json:"channel_id,omitempty"`
	VideoModelID            int64               `json:"video_model_id"`
	RequestedModel          string              `json:"requested_model"`
	UpstreamModel           string              `json:"upstream_model"`
	UpstreamTaskID          *string             `json:"upstream_task_id,omitempty"`
	Status                  string              `json:"status"`
	Progress                int                 `json:"progress"`
	BillingState            string              `json:"billing_state"`
	RequestPayload          map[string]any      `json:"request_payload,omitempty"`
	UpstreamRequestPayload  map[string]any      `json:"upstream_request_payload,omitempty"`
	UpstreamResponsePayload map[string]any      `json:"upstream_response_payload,omitempty"`
	ResultPayload           map[string]any      `json:"result_payload,omitempty"`
	ErrorCode               *string             `json:"error_code,omitempty"`
	ErrorMessage            *string             `json:"error_message,omitempty"`
	ContentURL              *string             `json:"content_url,omitempty"`
	UpstreamContentURL      *string             `json:"upstream_content_url,omitempty"`
	LocalContentURL         *string             `json:"local_content_url,omitempty"`
	BillingMode             string              `json:"billing_mode"`
	UnitPrice               float64             `json:"unit_price"`
	UnitSeconds             *float64            `json:"unit_seconds,omitempty"`
	RequestedSeconds        *int                `json:"requested_seconds,omitempty"`
	BillableSeconds         *int                `json:"billable_seconds,omitempty"`
	ReservedCost            float64             `json:"reserved_cost"`
	EstimatedCost           float64             `json:"estimated_cost"`
	ActualCost              float64             `json:"actual_cost"`
	SubmittedAt             *time.Time          `json:"submitted_at,omitempty"`
	StartedAt               *time.Time          `json:"started_at,omitempty"`
	CompletedAt             *time.Time          `json:"completed_at,omitempty"`
	ExpiresAt               *time.Time          `json:"expires_at,omitempty"`
	NextPollAt              *time.Time          `json:"next_poll_at,omitempty"`
	PollCount               int                 `json:"poll_count"`
	LockedUntil             *time.Time          `json:"locked_until,omitempty"`
	CreatedAt               time.Time           `json:"created_at"`
	UpdatedAt               time.Time           `json:"updated_at"`
	VideoModel              *videoModelResponse `json:"video_model,omitempty"`
}

func (h *VideoHandler) ListTemplates(c *gin.Context) {
	items, err := h.videoService.AdminListTemplates(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": videoTemplateResponses(items)})
}

func (h *VideoHandler) CreateTemplate(c *gin.Context) {
	var req videoTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.videoService.AdminCreateTemplate(c.Request.Context(), videoTemplateFromRequest(req))
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Created(c, videoTemplateResponseFromService(item))
}

func (h *VideoHandler) UpdateTemplate(c *gin.Context) {
	id, ok := parseVideoAdminID(c, "template")
	if !ok {
		return
	}
	var req videoTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.videoService.AdminUpdateTemplate(c.Request.Context(), id, videoTemplateFromRequest(req))
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, videoTemplateResponseFromService(item))
}

func (h *VideoHandler) DeleteTemplate(c *gin.Context) {
	id, ok := parseVideoAdminID(c, "template")
	if !ok {
		return
	}
	if err := h.videoService.AdminDeleteTemplate(c.Request.Context(), id); err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *VideoHandler) TestTemplateCreate(c *gin.Context) {
	var req videoTemplateCreateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.videoService.AdminTestTemplateCreate(c.Request.Context(), req.AccountID, req.TemplateID, req.Body)
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *VideoHandler) TestTemplateQuery(c *gin.Context) {
	var req videoTemplateQueryTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.videoService.AdminTestTemplateQuery(c.Request.Context(), req.AccountID, req.TemplateID, req.UpstreamTaskID)
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *VideoHandler) RecognizeTemplate(c *gin.Context) {
	var req videoTemplateRecognizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.videoService.AdminRecognizeTemplate(c.Request.Context(), req.AccountID, req.Model, req.Document)
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, videoTemplateResponseFromService(item))
}

func (h *VideoHandler) ListModels(c *gin.Context) {
	items, err := h.videoService.AdminListModels(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": videoModelResponses(items)})
}

func (h *VideoHandler) RequestShapes(c *gin.Context) {
	items, err := h.videoService.AdminRequestShapes(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

func (h *VideoHandler) CreateModel(c *gin.Context) {
	var req videoModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.videoService.AdminCreateModel(c.Request.Context(), videoModelFromRequest(req))
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Created(c, videoModelResponseFromService(item))
}

func (h *VideoHandler) UpdateModel(c *gin.Context) {
	id, ok := parseVideoAdminID(c, "model")
	if !ok {
		return
	}
	var req videoModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}
	item, err := h.videoService.AdminUpdateModel(c.Request.Context(), id, videoModelFromRequest(req))
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, videoModelResponseFromService(item))
}

func (h *VideoHandler) DeleteModel(c *gin.Context) {
	id, ok := parseVideoAdminID(c, "model")
	if !ok {
		return
	}
	if err := h.videoService.AdminDeleteModel(c.Request.Context(), id); err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *VideoHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	userID, _ := strconv.ParseInt(strings.TrimSpace(c.Query("user_id")), 10, 64)
	apiKeyID, _ := strconv.ParseInt(strings.TrimSpace(c.Query("api_key_id")), 10, 64)
	startAt, startOK := parseVideoAdminTime(c.Query("start_at"))
	if !startOK {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TIME_RANGE", "invalid start_at"))
		return
	}
	endAt, endOK := parseVideoAdminTime(c.Query("end_at"))
	if !endOK {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TIME_RANGE", "invalid end_at"))
		return
	}
	items, total, err := h.videoService.AdminListTasks(c.Request.Context(), service.VideoTaskFilter{
		Status:   strings.TrimSpace(c.Query("status")),
		Model:    strings.TrimSpace(c.Query("model")),
		UserID:   userID,
		APIKeyID: apiKeyID,
		StartAt:  startAt,
		EndAt:    endAt,
		Limit:    pageSize,
		Offset:   (page - 1) * pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"items":     videoTaskResponses(items),
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *VideoHandler) GetTask(c *gin.Context) {
	task, err := h.videoService.AdminGetTask(c.Request.Context(), strings.TrimSpace(c.Param("id")))
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, videoTaskResponseFromService(task))
}

func (h *VideoHandler) RequeueTask(c *gin.Context) {
	task, err := h.videoService.AdminRequeueTask(c.Request.Context(), strings.TrimSpace(c.Param("id")))
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, videoTaskResponseFromService(task))
}

func (h *VideoHandler) FailTask(c *gin.Context) {
	var req videoTaskActionRequest
	_ = c.ShouldBindJSON(&req)
	task, err := h.videoService.AdminFailTask(c.Request.Context(), strings.TrimSpace(c.Param("id")), req.Code, req.Message)
	if err != nil {
		h.writeVideoAdminError(c, err)
		return
	}
	response.Success(c, videoTaskResponseFromService(task))
}

func (h *VideoHandler) writeVideoAdminError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrVideoInvalidRequest):
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_VIDEO_CONFIG", err.Error()))
	case errors.Is(err, service.ErrVideoModelNotFound), errors.Is(err, service.ErrVideoTemplateNotFound):
		response.ErrorFrom(c, infraerrors.NotFound("VIDEO_CONFIG_NOT_FOUND", err.Error()))
	default:
		response.Error(c, http.StatusInternalServerError, "video config operation failed")
	}
}

func parseVideoAdminID(c *gin.Context, label string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_VIDEO_ID", "invalid "+label+" id"))
		return 0, false
	}
	return id, true
}

func videoTemplateFromRequest(req videoTemplateRequest) *service.VideoCallTemplate {
	return &service.VideoCallTemplate{
		Name:          req.Name,
		CreateMethod:  req.CreateMethod,
		CreatePath:    req.CreatePath,
		QueryMethod:   req.QueryMethod,
		QueryPath:     req.QueryPath,
		ContentMethod: req.ContentMethod,
		ContentPath:   req.ContentPath,
		CancelMethod:  req.CancelMethod,
		CancelPath:    req.CancelPath,
		StatusMapping: req.StatusMapping,
		ResultMapping: req.ResultMapping,
		ErrorMapping:  req.ErrorMapping,
		PollConfig:    req.PollConfig,
		TimeoutConfig: req.TimeoutConfig,
		Status:        req.Status,
	}
}

func videoModelFromRequest(req videoModelRequest) *service.VideoModel {
	return &service.VideoModel{
		PublicModel:      req.PublicModel,
		DisplayName:      req.DisplayName,
		TemplateID:       req.TemplateID,
		UpstreamModelID:  req.UpstreamModelID,
		RequestShape:     req.RequestShape,
		Status:           req.Status,
		Capabilities:     req.Capabilities,
		Defaults:         req.Defaults,
		Limits:           req.Limits,
		SupportedOptions: req.SupportedOptions,
		ExtraBodyAllow:   req.ExtraBodyAllow,
		SortOrder:        req.SortOrder,
	}
}

func videoTemplateResponses(items []service.VideoCallTemplate) []videoTemplateResponse {
	out := make([]videoTemplateResponse, 0, len(items))
	for i := range items {
		out = append(out, *videoTemplateResponseFromService(&items[i]))
	}
	return out
}

func videoTemplateResponseFromService(item *service.VideoCallTemplate) *videoTemplateResponse {
	if item == nil {
		return nil
	}
	return &videoTemplateResponse{
		ID:            item.ID,
		Name:          item.Name,
		CreateMethod:  item.CreateMethod,
		CreatePath:    item.CreatePath,
		QueryMethod:   item.QueryMethod,
		QueryPath:     item.QueryPath,
		ContentMethod: item.ContentMethod,
		ContentPath:   item.ContentPath,
		CancelMethod:  item.CancelMethod,
		CancelPath:    item.CancelPath,
		StatusMapping: item.StatusMapping,
		ResultMapping: item.ResultMapping,
		ErrorMapping:  item.ErrorMapping,
		PollConfig:    item.PollConfig,
		TimeoutConfig: item.TimeoutConfig,
		Status:        item.Status,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func videoModelResponses(items []service.VideoModel) []videoModelResponse {
	out := make([]videoModelResponse, 0, len(items))
	for i := range items {
		out = append(out, *videoModelResponseFromService(&items[i]))
	}
	return out
}

func videoModelResponseFromService(item *service.VideoModel) *videoModelResponse {
	if item == nil {
		return nil
	}
	return &videoModelResponse{
		ID:               item.ID,
		PublicModel:      item.PublicModel,
		DisplayName:      item.DisplayName,
		TemplateID:       item.TemplateID,
		UpstreamModelID:  item.UpstreamModelID,
		RequestShape:     item.RequestShape,
		Status:           item.Status,
		Capabilities:     item.Capabilities,
		Defaults:         item.Defaults,
		Limits:           item.Limits,
		SupportedOptions: item.SupportedOptions,
		ExtraBodyAllow:   item.ExtraBodyAllow,
		SortOrder:        item.SortOrder,
		Template:         videoTemplateResponseFromService(item.Template),
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func videoTaskResponses(items []service.VideoGenerationTask) []videoTaskResponse {
	out := make([]videoTaskResponse, 0, len(items))
	for i := range items {
		out = append(out, *videoTaskResponseFromService(&items[i]))
	}
	return out
}

func videoTaskResponseFromService(item *service.VideoGenerationTask) *videoTaskResponse {
	if item == nil {
		return nil
	}
	return &videoTaskResponse{
		ID:                      item.ID,
		PublicID:                item.PublicID,
		UserID:                  item.UserID,
		APIKeyID:                item.APIKeyID,
		GroupID:                 item.GroupID,
		AccountID:               item.AccountID,
		ChannelID:               item.ChannelID,
		VideoModelID:            item.VideoModelID,
		RequestedModel:          item.RequestedModel,
		UpstreamModel:           item.UpstreamModel,
		UpstreamTaskID:          item.UpstreamTaskID,
		Status:                  item.Status,
		Progress:                item.Progress,
		BillingState:            item.BillingState,
		RequestPayload:          item.RequestPayload,
		UpstreamRequestPayload:  item.UpstreamRequestPayload,
		UpstreamResponsePayload: item.UpstreamResponsePayload,
		ResultPayload:           item.ResultPayload,
		ErrorCode:               item.ErrorCode,
		ErrorMessage:            item.ErrorMessage,
		ContentURL:              item.ContentURL,
		UpstreamContentURL:      item.UpstreamContentURL,
		LocalContentURL:         item.LocalContentURL,
		BillingMode:             item.BillingMode,
		UnitPrice:               item.UnitPrice,
		UnitSeconds:             item.UnitSeconds,
		RequestedSeconds:        item.RequestedSeconds,
		BillableSeconds:         item.BillableSeconds,
		ReservedCost:            item.ReservedCost,
		EstimatedCost:           item.EstimatedCost,
		ActualCost:              item.ActualCost,
		SubmittedAt:             item.SubmittedAt,
		StartedAt:               item.StartedAt,
		CompletedAt:             item.CompletedAt,
		ExpiresAt:               item.ExpiresAt,
		NextPollAt:              item.NextPollAt,
		PollCount:               item.PollCount,
		LockedUntil:             item.LockedUntil,
		CreatedAt:               item.CreatedAt,
		UpdatedAt:               item.UpdatedAt,
		VideoModel:              videoModelResponseFromService(item.VideoModel),
	}
}

func parseVideoAdminTime(raw string) (*time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, false
	}
	return &t, true
}
