package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/Wei-Shaw/ccapi/internal/service"
)

// S3BackupStore implements service.BackupObjectStore using AWS S3 compatible storage
type S3BackupStore struct {
	client *s3.Client
	bucket string
}

// NewS3BackupStoreFactory returns a BackupObjectStoreFactory that creates S3-backed stores
func NewS3BackupStoreFactory() service.BackupObjectStoreFactory {
	return func(ctx context.Context, cfg *service.BackupS3Config) (service.BackupObjectStore, error) {
		if strings.EqualFold(strings.TrimSpace(cfg.Provider), "local") {
			return NewLocalBackupStore(cfg.LocalPath, cfg.PublicBaseURL)
		}

		region := cfg.Region
		if region == "" {
			region = "auto" // Cloudflare R2 默认 region
		}

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("load aws config: %w", err)
		}

		client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if cfg.Endpoint != "" {
				o.BaseEndpoint = &cfg.Endpoint
			}
			if cfg.ForcePathStyle {
				o.UsePathStyle = true
			}
			o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
			o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		})

		return &S3BackupStore{client: client, bucket: cfg.Bucket}, nil
	}
}

// LocalBackupStore implements BackupObjectStore using the local filesystem.
type LocalBackupStore struct {
	root          string
	publicBaseURL string
}

func NewLocalBackupStore(root string, publicBaseURL string) (*LocalBackupStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("local storage path is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve local storage path: %w", err)
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("create local storage path: %w", err)
	}
	return &LocalBackupStore{
		root:          abs,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
	}, nil
}

func (s *LocalBackupStore) Upload(ctx context.Context, key string, body io.Reader, contentType string) (int64, error) {
	path, err := s.pathForKey(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return 0, fmt.Errorf("create local object directory: %w", err)
	}
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return 0, fmt.Errorf("create local object: %w", err)
	}
	written, copyErr := io.Copy(file, body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("write local object: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("close local object: %w", closeErr)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return written, fmt.Errorf("commit local object: %w", err)
	}
	return written, ctx.Err()
}

func (s *LocalBackupStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.pathForKey(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open local object: %w", err)
	}
	return file, nil
}

func (s *LocalBackupStore) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete local object: %w", err)
	}
	return nil
}

func (s *LocalBackupStore) PresignURL(ctx context.Context, key string, _ time.Duration) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + escapeObjectKey(key), nil
	}
	path, err := s.pathForKey(key)
	if err != nil {
		return "", err
	}
	return "file://" + path, nil
}

func (s *LocalBackupStore) HeadBucket(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(s.root)
	if err != nil {
		return fmt.Errorf("stat local storage path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("local storage path is not a directory")
	}
	testFile := filepath.Join(s.root, ".ccapi-write-test")
	if err := os.WriteFile(testFile, []byte("ok"), 0o640); err != nil {
		return fmt.Errorf("write local storage test file: %w", err)
	}
	_ = os.Remove(testFile)
	return nil
}

func (s *LocalBackupStore) pathForKey(key string) (string, error) {
	cleanKey := filepath.Clean(strings.TrimLeft(strings.TrimSpace(key), "/"))
	if cleanKey == "." || cleanKey == "" || strings.HasPrefix(cleanKey, "..") {
		return "", fmt.Errorf("invalid local object key")
	}
	path := filepath.Join(s.root, cleanKey)
	rel, err := filepath.Rel(s.root, path)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid local object key")
	}
	return path, nil
}

func escapeObjectKey(key string) string {
	parts := strings.Split(strings.TrimLeft(key, "/"), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func (s *S3BackupStore) Upload(ctx context.Context, key string, body io.Reader, contentType string) (int64, error) {
	// 读取全部内容以获取大小（S3 PutObject 需要知道内容长度）
	// 注意：阿里云 OSS 不兼容 s3manager 分片上传的签名方式，因此使用 PutObject
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, fmt.Errorf("read body: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return 0, fmt.Errorf("S3 PutObject: %w", err)
	}
	return int64(len(data)), nil
}

func (s *S3BackupStore) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("S3 GetObject: %w", err)
	}
	return result.Body, nil
}

func (s *S3BackupStore) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	return err
}

func (s *S3BackupStore) PresignURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign url: %w", err)
	}
	return result.URL, nil
}

func (s *S3BackupStore) HeadBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &s.bucket,
	})
	if err != nil {
		return fmt.Errorf("S3 HeadBucket failed: %w", err)
	}
	return nil
}
