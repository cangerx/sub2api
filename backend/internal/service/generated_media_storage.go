package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type generatedMediaStore struct {
	settingService *SettingService
	storeFactory   BackupObjectStoreFactory
}

type generatedImagePayload struct {
	Data         string
	MIMEType     string
	OutputFormat string
}

func (s generatedMediaStore) persistImages(ctx context.Context, c *gin.Context, provider string, images []generatedImagePayload) ([]string, error) {
	if len(images) == 0 || s.settingService == nil || s.storeFactory == nil {
		return nil, nil
	}
	cfg, err := s.settingService.GetBackupStorageConfig(ctx)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if cfg == nil || !cfg.IsConfigured() {
		return nil, nil
	}
	if normalizeBackupStorageProvider(cfg.Provider) == "local" && strings.TrimSpace(cfg.PublicBaseURL) == "" {
		clone := *cfg
		clone.PublicBaseURL = generatedMediaLocalPublicBaseURL(c)
		cfg = &clone
	}
	store, err := s.storeFactory(ctx, cfg)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(images))
	for _, img := range images {
		raw := strings.TrimSpace(img.Data)
		if raw == "" {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("decode generated image: %w", err)
		}
		sum := sha256.Sum256(data)
		format := strings.TrimSpace(img.OutputFormat)
		if format == "" {
			format = strings.TrimSpace(img.MIMEType)
		}
		key := fmt.Sprintf(
			"images/%s/%s/%s-%s%s",
			safeGeneratedMediaProvider(provider),
			time.Now().UTC().Format("2006/01/02"),
			uuid.NewString(),
			hex.EncodeToString(sum[:8]),
			openAIImageOutputExtension(format),
		)
		contentType := generatedImageMIMEType(img.MIMEType, format)
		if _, err := store.Upload(ctx, key, bytes.NewReader(data), contentType); err != nil {
			return nil, fmt.Errorf("store generated image: %w", err)
		}
		mediaURL, err := store.PresignURL(ctx, key, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("build generated image URL: %w", err)
		}
		if trimmed := strings.TrimSpace(mediaURL); trimmed != "" {
			urls = append(urls, trimmed)
		}
	}
	return urls, nil
}

func generatedMediaLocalPublicBaseURL(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return "/api/v1/media"
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = strings.TrimSpace(strings.Split(proto, ",")[0])
	}
	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = c.Request.Host
	}
	if host == "" {
		return "/api/v1/media"
	}
	return strings.TrimRight(scheme+"://"+host, "/") + "/api/v1/media"
}

func generatedImageMIMEType(mimeType string, outputFormat string) string {
	if trimmed := strings.TrimSpace(mimeType); trimmed != "" {
		return trimmed
	}
	if detected := openAIImageOutputMIMEType(outputFormat); detected != "" {
		return detected
	}
	return "image/png"
}

func safeGeneratedMediaProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return "generated"
	}
	var b strings.Builder
	for _, r := range provider {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "generated"
	}
	return b.String()
}

func collectGeminiInlineImagePayloads(resp map[string]any) []generatedImagePayload {
	parts := extractGeminiParts(resp)
	if len(parts) == 0 {
		return nil
	}
	images := make([]generatedImagePayload, 0)
	for _, part := range parts {
		inlineData, ok := part["inlineData"].(map[string]any)
		if !ok {
			continue
		}
		data, _ := inlineData["data"].(string)
		mimeType, _ := inlineData["mimeType"].(string)
		if strings.TrimSpace(data) == "" {
			continue
		}
		images = append(images, generatedImagePayload{
			Data:     data,
			MIMEType: mimeType,
		})
	}
	return images
}
