package repository

import (
	"context"
	"log"

	"github.com/Wei-Shaw/ccapi/ent"
	"github.com/Wei-Shaw/ccapi/ent/videocalltemplate"
)

// ensureBuiltinVideoTemplates 在启动时将官方内置视频接口模板写入数据库。
// 幂等：ON CONFLICT (name) DO NOTHING，不会覆盖管理员已编辑的同名模板。
// 三家提供商：OpenAI 风格、Google Gemini/Veo、火山 Ark。
func ensureBuiltinVideoTemplates(ctx context.Context, client *ent.Client) error {
	// 通用状态映射（大多数提供商共用）
	baseStatusMapping := map[string]string{
		"queued":      "queued",
		"pending":     "queued",
		"created":     "queued",
		"processing":  "in_progress",
		"running":     "in_progress",
		"in_progress": "in_progress",
		"submitted":   "in_progress",
		"succeeded":   "completed",
		"success":     "completed",
		"succeed":     "completed",
		"completed":   "completed",
		"done":        "completed",
		"failed":      "failed",
		"error":       "failed",
		"cancelled":   "cancelled",
		"canceled":    "cancelled",
		"expired":     "expired",
	}
	basePollConfig := map[string]interface{}{
		"interval_seconds":    5,
		"backoff_max_seconds": 30,
		"max_attempts":        240,
	}
	baseTimeoutConfig := map[string]interface{}{
		"create_seconds":  60,
		"query_seconds":   30,
		"content_seconds": 300,
	}

	type tplDef struct {
		name          string
		createMethod  string
		createPath    string
		queryMethod   string
		queryPath     string
		contentMethod *string
		contentPath   *string
		cancelMethod  *string
		cancelPath    *string
		statusMapping map[string]string
		resultMapping map[string]string
		errorMapping  map[string]string
		pollConfig    map[string]interface{}
		timeoutConfig map[string]interface{}
	}

	strPtr := func(s string) *string { return &s }

	templates := []tplDef{
		{
			// OpenAI 风格异步视频接口（Sub2API 自身、以及与 OpenAI 接口对齐的中转服务）
			name:          "内置 - OpenAI 风格视频接口",
			createMethod:  "POST",
			createPath:    "/v1/videos",
			queryMethod:   "GET",
			queryPath:     "/v1/videos/{task_id}",
			contentMethod: strPtr("GET"),
			contentPath:   strPtr("/v1/videos/{task_id}/content.mp4"),
			cancelMethod:  strPtr("POST"),
			cancelPath:    strPtr("/v1/videos/{task_id}/cancel"),
			statusMapping: baseStatusMapping,
			resultMapping: map[string]string{
				"content_url": "content_url",
				"seconds":     "seconds",
				"progress":    "progress",
			},
			errorMapping: map[string]string{
				"code":    "error.code",
				"message": "error.message",
			},
			pollConfig:    basePollConfig,
			timeoutConfig: baseTimeoutConfig,
		},
		{
			// Google Gemini / Veo 官方视频生成接口
			// Create: POST /v1beta/models/{model}:predictLongRunning
			// Query:  GET  /v1beta/{task_id}  (operation name 即 task_id)
			// Cancel: POST /v1beta/{task_id}:cancel
			name:          "内置 - Google Gemini / Veo 官方视频",
			createMethod:  "POST",
			createPath:    "/v1beta/models/{model}:predictLongRunning",
			queryMethod:   "GET",
			queryPath:     "/v1beta/{task_id}",
			contentMethod: nil,
			contentPath:   nil,
			cancelMethod:  strPtr("POST"),
			cancelPath:    strPtr("/v1beta/{task_id}:cancel"),
			statusMapping: func() map[string]string {
				m := make(map[string]string, len(baseStatusMapping)+1)
				for k, v := range baseStatusMapping {
					m[k] = v
				}
				m["done"] = "completed" // Gemini operation.done=true → completed
				return m
			}(),
			resultMapping: map[string]string{
				"content_url": "response.generateVideoResponse.generatedSamples.0.video.uri",
				"seconds":     "response.generateVideoResponse.generatedSamples.0.video.duration",
				"progress":    "metadata.progressPercentage",
			},
			errorMapping: map[string]string{
				"code":    "error.code",
				"message": "error.message",
			},
			pollConfig: map[string]interface{}{
				"interval_seconds":    10,
				"backoff_max_seconds": 60,
				"max_attempts":        180,
			},
			timeoutConfig: baseTimeoutConfig,
		},
		{
			// 火山引擎 Ark / 豆包官方视频生成接口
			// Create: POST /api/v3/contents/generations/tasks
			// Query:  GET  /api/v3/contents/generations/tasks/{task_id}
			name:          "内置 - 火山 Ark / 豆包官方视频",
			createMethod:  "POST",
			createPath:    "/api/v3/contents/generations/tasks",
			queryMethod:   "GET",
			queryPath:     "/api/v3/contents/generations/tasks/{task_id}",
			contentMethod: nil,
			contentPath:   nil,
			cancelMethod:  nil,
			cancelPath:    nil,
			statusMapping: baseStatusMapping,
			resultMapping: map[string]string{
				"content_url": "data.video_url",
				"seconds":     "data.duration",
				"progress":    "data.progress",
			},
			errorMapping: map[string]string{
				"code":    "error.code",
				"message": "error.message",
			},
			pollConfig:    basePollConfig,
			timeoutConfig: baseTimeoutConfig,
		},
		{
			// MegaByAI 异步视频接口（/v1/videos 路径，content 无 .mp4 后缀）
			name:          "MegaByAI - 异步视频接口",
			createMethod:  "POST",
			createPath:    "/v1/videos",
			queryMethod:   "GET",
			queryPath:     "/v1/videos/{task_id}",
			contentMethod: strPtr("GET"),
			contentPath:   strPtr("/v1/videos/{task_id}/content"),
			cancelMethod:  nil,
			cancelPath:    nil,
			statusMapping: map[string]string{
				"queued":      "queued",
				"in_progress": "in_progress",
				"completed":   "completed",
				"failed":      "failed",
			},
			resultMapping: map[string]string{
				"content_url": "video_url|url|metadata.content_url",
				"seconds":     "seconds",
				"progress":    "progress",
			},
			errorMapping: map[string]string{
				"code":    "error.code",
				"message": "error.message",
			},
			pollConfig: map[string]interface{}{
				"interval_seconds":    5,
				"backoff_max_seconds": 10,
				"max_attempts":        180,
			},
			timeoutConfig: baseTimeoutConfig,
		},
		{
			// 多米API SEEDANCE 视频（火山 Ark 路径格式，result 在 data.videos）
			name:          "多米API - SEEDANCE 视频",
			createMethod:  "POST",
			createPath:    "/api/v3/contents/generations/tasks",
			queryMethod:   "GET",
			queryPath:     "/api/v3/contents/generations/tasks/{task_id}",
			contentMethod: nil,
			contentPath:   nil,
			cancelMethod:  nil,
			cancelPath:    nil,
			statusMapping: baseStatusMapping,
			resultMapping: map[string]string{
				"content_url": "data.videos.0.url",
				"seconds":     "data.duration",
				"progress":    "progress",
			},
			errorMapping: map[string]string{
				"code":    "state",
				"message": "message",
			},
			pollConfig:    basePollConfig,
			timeoutConfig: baseTimeoutConfig,
		},
		{
			// 多米API KLING 文生视频
			name:          "多米API - KLING 文生视频",
			createMethod:  "POST",
			createPath:    "/api/video/kling/v1/videos/text2video",
			queryMethod:   "GET",
			queryPath:     "/api/video/kling/v1/videos/text2video/{task_id}",
			contentMethod: nil,
			contentPath:   nil,
			cancelMethod:  nil,
			cancelPath:    nil,
			statusMapping: baseStatusMapping,
			resultMapping: map[string]string{
				"content_url": "data.task_result.videos.0.url",
				"seconds":     "data.task_result.videos.0.duration",
				"progress":    "data.progress",
			},
			errorMapping: map[string]string{
				"code":    "data.task_status",
				"message": "data.task_status_msg",
			},
			pollConfig:    basePollConfig,
			timeoutConfig: baseTimeoutConfig,
		},
		{
			// 多米API 异步图片生成（gpt-image-2 / NANO-BANANA）
			name:          "多米API - 图片生成（gpt-image-2）",
			createMethod:  "POST",
			createPath:    "/v1/images/generations?async=true",
			queryMethod:   "GET",
			queryPath:     "/v1/tasks/{task_id}",
			contentMethod: nil,
			contentPath:   nil,
			cancelMethod:  nil,
			cancelPath:    nil,
			statusMapping: baseStatusMapping,
			resultMapping: map[string]string{
				"content_url": "data.images.0.url",
				"seconds":     "",
				"progress":    "progress",
			},
			errorMapping: map[string]string{
				"code":    "state",
				"message": "message",
			},
			pollConfig:    basePollConfig,
			timeoutConfig: baseTimeoutConfig,
		},
	}

	seeded := 0
	for _, t := range templates {
		// 跳过库里已有同名的模板（保护管理员手动修改过的同名记录）
		exists, err := client.VideoCallTemplate.
			Query().
			Where(videocalltemplate.NameEQ(t.name)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		c := client.VideoCallTemplate.
			Create().
			SetName(t.name).
			SetCreateMethod(t.createMethod).
			SetCreatePath(t.createPath).
			SetQueryMethod(t.queryMethod).
			SetQueryPath(t.queryPath).
			SetNillableContentMethod(t.contentMethod).
			SetNillableContentPath(t.contentPath).
			SetNillableCancelMethod(t.cancelMethod).
			SetNillableCancelPath(t.cancelPath).
			SetStatusMapping(t.statusMapping).
			SetResultMapping(t.resultMapping).
			SetErrorMapping(t.errorMapping).
			SetPollConfig(t.pollConfig).
			SetTimeoutConfig(t.timeoutConfig).
			SetStatus("active")

		if _, err := c.Save(ctx); err != nil {
			return err
		}
		seeded++
	}

	if seeded > 0 {
		log.Printf("video template bootstrap: seeded %d built-in template(s)", seeded)
	}
	return nil
}
