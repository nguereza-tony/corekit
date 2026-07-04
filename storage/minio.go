package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nguereza-tony/corekit/config"
	"github.com/nguereza-tony/corekit/logger"
)

type MinIOStorage struct {
	client *minio.Client
	bucket string
	logger *logger.Logger
}

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(cfg *config.StorageConfig, log *logger.Logger) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Verify bucket exists (bucket must be created manually by admin)
	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket '%s' does not exist. Please create it first", cfg.Bucket)
	}

	log.Infof("Connected to MinIO bucket: %s", cfg.Bucket)

	return &MinIOStorage{
		client: client,
		bucket: cfg.Bucket,
		logger: log,
	}, nil
}

// Put uploads a file to the bucket
func (s *MinIOStorage) Put(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, path, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Get downloads a file from the bucket
func (s *MinIOStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Delete removes a file from the bucket
func (s *MinIOStorage) Delete(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}

// GetSignedURL returns a presigned URL for temporary access
func (s *MinIOStorage) GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, path, expiry, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// Exists checks if a file exists in the bucket
func (s *MinIOStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBucket returns the bucket name
func (s *MinIOStorage) GetBucket() string {
	return s.bucket
}
