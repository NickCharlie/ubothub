package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo holds metadata about a stored object.
type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// ObjectStorage defines the interface for object storage operations.
// Implementations must be safe for concurrent use.
type ObjectStorage interface {
	// EnsureBucket creates the bucket if it does not exist.
	EnsureBucket(ctx context.Context, bucket string) error

	// PutObject uploads a file to the specified bucket and key.
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error

	// GetObject retrieves a file from the specified bucket and key.
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// DeleteObject removes a file from the specified bucket and key.
	DeleteObject(ctx context.Context, bucket, key string) error

	// PresignedPutURL generates a pre-signed URL for direct client upload.
	PresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)

	// PresignedGetURL generates a pre-signed URL for direct client download.
	PresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)

	// ObjectExists checks whether an object exists at the given key.
	ObjectExists(ctx context.Context, bucket, key string) (bool, error)

	// StatObject returns metadata about the object at the given key.
	StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)
}
