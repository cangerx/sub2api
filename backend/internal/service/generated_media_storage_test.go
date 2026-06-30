//go:build unit

package service

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/ccapi/internal/config"
	"github.com/stretchr/testify/require"
)

type generatedMediaObjectStoreStub struct {
	uploadedKeys []string
	contentTypes []string
}

func (s *generatedMediaObjectStoreStub) Upload(_ context.Context, key string, body io.Reader, contentType string) (int64, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, err
	}
	s.uploadedKeys = append(s.uploadedKeys, key)
	s.contentTypes = append(s.contentTypes, contentType)
	return int64(len(data)), nil
}

func (s *generatedMediaObjectStoreStub) Download(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (s *generatedMediaObjectStoreStub) Delete(context.Context, string) error { return nil }

func (s *generatedMediaObjectStoreStub) PresignURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://cdn.example.com/" + key, nil
}

func (s *generatedMediaObjectStoreStub) HeadBucket(context.Context) error { return nil }

func TestGeneratedMediaStorePersistsGeminiInlineImages(t *testing.T) {
	cfg := BackupS3Config{
		Provider:        "s3",
		Bucket:          "media",
		AccessKeyID:     "ak",
		SecretAccessKey: "ENC:sk",
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	settingSvc := NewSettingService(&settingRepoStub{
		values: map[string]string{settingKeyBackupS3Config: string(raw)},
	}, &config.Config{})
	settingSvc.SetSecretEncryptor(&plainEncryptor{})
	store := &generatedMediaObjectStoreStub{}
	var receivedSecret string
	media := generatedMediaStore{
		settingService: settingSvc,
		storeFactory: func(_ context.Context, cfg *BackupS3Config) (BackupObjectStore, error) {
			receivedSecret = cfg.SecretAccessKey
			return store, nil
		},
	}

	resp := map[string]any{
		"candidates": []any{
			map[string]any{
				"content": map[string]any{
					"parts": []any{
						map[string]any{
							"inlineData": map[string]any{
								"mimeType": "image/png",
								"data":     "aGVsbG8=",
							},
						},
					},
				},
			},
		},
	}

	urls, err := media.persistImages(context.Background(), nil, "gemini", collectGeminiInlineImagePayloads(resp))
	require.NoError(t, err)
	require.Len(t, urls, 1)
	require.Contains(t, urls[0], "https://cdn.example.com/images/gemini/")
	require.Len(t, store.uploadedKeys, 1)
	require.Contains(t, store.uploadedKeys[0], "images/gemini/")
	require.Equal(t, "image/png", store.contentTypes[0])
	require.Equal(t, "sk", receivedSecret)
}
