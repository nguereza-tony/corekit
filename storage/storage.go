package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for object storage operations
type Storage interface {
	// Put uploads a file to the bucket
	Put(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error

	// Get downloads a file from the bucket
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes a file from the bucket
	Delete(ctx context.Context, path string) error

	// GetSignedURL returns a presigned URL for temporary access
	GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// Exists checks if a file exists in the bucket
	Exists(ctx context.Context, path string) (bool, error)

	// GetBucket returns the bucket name
	GetBucket() string
}
