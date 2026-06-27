package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// ResolveLocalMediaPath resolves an object key against the configured local
// storage root. It is used by the public media route for generated assets.
func (s *SettingService) ResolveLocalMediaPath(ctx context.Context, key string) (string, error) {
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if key == "" || strings.Contains(key, "\x00") {
		return "", fmt.Errorf("invalid media key")
	}
	cfg, err := s.GetBackupStorageConfig(ctx)
	if err != nil {
		return "", err
	}
	if normalizeBackupStorageProvider(cfg.Provider) != "local" || strings.TrimSpace(cfg.LocalPath) == "" {
		return "", ErrBackupS3NotConfigured
	}
	root, err := filepath.Abs(cfg.LocalPath)
	if err != nil {
		return "", fmt.Errorf("resolve local storage path: %w", err)
	}
	full := filepath.Join(root, filepath.FromSlash(key))
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", fmt.Errorf("invalid media key")
	}
	return full, nil
}
