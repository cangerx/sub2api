package repository

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalBackupStoreUploadDownloadDelete(t *testing.T) {
	store, err := NewLocalBackupStore(t.TempDir(), "https://cdn.example.com/files")
	require.NoError(t, err)

	size, err := store.Upload(context.Background(), "backups/2026/test.sql.gz", strings.NewReader("backup-data"), "application/gzip")
	require.NoError(t, err)
	require.Equal(t, int64(len("backup-data")), size)

	reader, err := store.Download(context.Background(), "backups/2026/test.sql.gz")
	require.NoError(t, err)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, "backup-data", string(data))

	url, err := store.PresignURL(context.Background(), "backups/2026/test.sql.gz", time.Hour)
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example.com/files/backups/2026/test.sql.gz", url)

	require.NoError(t, store.HeadBucket(context.Background()))
	require.NoError(t, store.Delete(context.Background(), "backups/2026/test.sql.gz"))
	_, err = store.Download(context.Background(), "backups/2026/test.sql.gz")
	require.Error(t, err)
}
