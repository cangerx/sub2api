package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	pkghttputil "github.com/Wei-Shaw/ccapi/internal/pkg/httputil"
	middleware2 "github.com/Wei-Shaw/ccapi/internal/server/middleware"
	"github.com/Wei-Shaw/ccapi/internal/service"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoService *service.VideoService
}

func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{videoService: videoService}
}

func (h *VideoHandler) Create(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey.User == nil {
		videoError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}

	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		if maxErr, ok := extractMaxBytesError(err); ok {
			videoError(c, http.StatusRequestEntityTooLarge, "invalid_request_error", buildBodyTooLargeMessage(maxErr.Limit))
			return
		}
		videoError(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(body) == 0 {
		videoError(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	req, err := service.ParseVideoCreateRequest(body)
	if err != nil {
		videoError(c, http.StatusBadRequest, "invalid_request_error", "Failed to parse request body")
		return
	}
	if err := h.videoService.ValidateVideoURLs(req); err != nil {
		videoError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
		return
	}

	input := service.VideoCreateInput{
		Request:        req,
		User:           apiKey.User,
		APIKey:         apiKey,
		GroupID:        apiKey.GroupID,
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
		UserAgent:      c.GetHeader("User-Agent"),
		IPAddress:      c.ClientIP(),
	}
	task, replayed, err := h.createIdempotent(c, input)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}
	if replayed {
		c.Header("X-Idempotency-Replayed", "true")
	}
	c.JSON(http.StatusAccepted, service.TaskToVideoObject(task))
}

func (h *VideoHandler) createIdempotent(c *gin.Context, input service.VideoCreateInput) (*service.VideoGenerationTask, bool, error) {
	coordinator := service.DefaultIdempotencyCoordinator()
	if coordinator == nil || strings.TrimSpace(input.IdempotencyKey) == "" {
		task, err := h.videoService.Create(c.Request.Context(), input)
		return task, false, err
	}
	actorScope := "apikey:" + strconv.FormatInt(input.APIKey.ID, 10)
	result, err := coordinator.Execute(c.Request.Context(), service.IdempotencyExecuteOptions{
		Scope:          "gateway.videos.create",
		ActorScope:     actorScope,
		Method:         http.MethodPost,
		Route:          "/v1/videos",
		IdempotencyKey: input.IdempotencyKey,
		Payload:        input.Request.Raw,
		RequireKey:     false,
		TTL:            24 * time.Hour,
	}, func(ctx context.Context) (any, error) {
		task, createErr := h.videoService.Create(ctx, input)
		if createErr != nil {
			return nil, createErr
		}
		return service.TaskToVideoObject(task), nil
	})
	if err != nil {
		return nil, false, err
	}
	if obj, ok := result.Data.(service.VideoObject); ok && obj.ID != "" {
		task, getErr := h.videoService.Get(c.Request.Context(), obj.ID, input.APIKey.ID)
		return task, result.Replayed, getErr
	}
	if raw, ok := result.Data.(map[string]any); ok {
		if id, _ := raw["id"].(string); id != "" {
			task, getErr := h.videoService.Get(c.Request.Context(), id, input.APIKey.ID)
			return task, result.Replayed, getErr
		}
	}
	return nil, result.Replayed, service.ErrVideoInvalidRequest
}

func (h *VideoHandler) Get(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	task, err := h.videoService.Get(c.Request.Context(), c.Param("id"), apiKey.ID)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, service.TaskToVideoObject(task))
}

func (h *VideoHandler) List(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	tasks, err := h.videoService.List(c.Request.Context(), apiKey.ID, limit, strings.TrimSpace(c.Query("after")))
	if err != nil {
		h.handleVideoError(c, err)
		return
	}
	resp := service.VideoListResponse{Object: "list", Data: make([]service.VideoObject, 0, len(tasks))}
	for i := range tasks {
		task := tasks[i]
		resp.Data = append(resp.Data, service.TaskToVideoObject(&task))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *VideoHandler) Cancel(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	task, err := h.videoService.Cancel(c.Request.Context(), c.Param("id"), apiKey.ID)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, service.TaskToVideoObject(task))
}

func (h *VideoHandler) Models(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	resp, err := h.videoService.ListModels(c.Request.Context(), apiKey.GroupID)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *VideoHandler) Content(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoError(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	resp, err := h.videoService.ContentResponse(c.Request.Context(), c.Param("id"), apiKey.ID)
	if err != nil {
		h.handleVideoError(c, err)
		return
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", `attachment; filename="`+c.Param("id")+`.mp4"`)
	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func (h *VideoHandler) handleVideoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrVideoModelNotFound), errors.Is(err, service.ErrVideoTaskNotFound):
		videoError(c, http.StatusNotFound, "not_found_error", err.Error())
	case errors.Is(err, service.ErrVideoModelDisabled):
		videoError(c, http.StatusForbidden, "permission_error", err.Error())
	case errors.Is(err, service.ErrVideoInvalidRequest):
		videoError(c, http.StatusBadRequest, "invalid_request_error", err.Error())
	case errors.Is(err, service.ErrVideoInsufficientBalance):
		videoError(c, http.StatusForbidden, "insufficient_balance", err.Error())
	case errors.Is(err, service.ErrVideoPricingUnavailable):
		videoError(c, http.StatusServiceUnavailable, "pricing_unavailable", err.Error())
	case errors.Is(err, service.ErrVideoContentUnavailable):
		videoError(c, http.StatusConflict, "content_unavailable", err.Error())
	case errors.Is(err, service.ErrVideoContentExpired):
		videoError(c, http.StatusGone, "content_expired", err.Error())
	case errors.Is(err, service.ErrNoAvailableAccounts):
		videoError(c, http.StatusServiceUnavailable, "no_available_accounts", err.Error())
	default:
		videoError(c, http.StatusInternalServerError, "api_error", "Internal server error")
	}
}

func videoError(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}
