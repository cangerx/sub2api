//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/ccapi/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSettingServiceGetBackupStorageConfigDecryptsRemoteSecret(t *testing.T) {
	cfg := BackupS3Config{
		Provider:        "s3",
		Bucket:          "media",
		AccessKeyID:     "ak",
		SecretAccessKey: "ENC:sk",
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	svc := NewSettingService(&settingRepoStub{
		values: map[string]string{settingKeyBackupS3Config: string(raw)},
	}, &config.Config{})
	svc.SetSecretEncryptor(&plainEncryptor{})

	got, err := svc.GetBackupStorageConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "sk", got.SecretAccessKey)
}

func TestSettingServiceGetBackupStorageConfigPreservesLegacyPlaintextSecret(t *testing.T) {
	cfg := BackupS3Config{
		Provider:        "s3",
		Bucket:          "media",
		AccessKeyID:     "ak",
		SecretAccessKey: "legacy-sk",
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	svc := NewSettingService(&settingRepoStub{
		values: map[string]string{settingKeyBackupS3Config: string(raw)},
	}, &config.Config{})
	svc.SetSecretEncryptor(&plainEncryptor{})

	got, err := svc.GetBackupStorageConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "legacy-sk", got.SecretAccessKey)
}

func TestSettingServiceGetBackupStorageConfigLeavesLocalSecretUntouched(t *testing.T) {
	cfg := BackupS3Config{
		Provider:        "local",
		LocalPath:       "./data/storage",
		SecretAccessKey: "ENC:unused",
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	svc := NewSettingService(&settingRepoStub{
		values: map[string]string{settingKeyBackupS3Config: string(raw)},
	}, &config.Config{})
	svc.SetSecretEncryptor(&plainEncryptor{})

	got, err := svc.GetBackupStorageConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "ENC:unused", got.SecretAccessKey)
}
