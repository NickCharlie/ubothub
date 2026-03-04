package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements ObjectStorage using MinIO as the backing store.
type MinIOStorage struct {
	client           *minio.Client
	externalEndpoint string // If set, presigned URLs are rewritten to this host.
	useSSL           bool
}

// MinIOConfig holds connection parameters for MinIO.
type MinIOConfig struct {
	Endpoint         string
	ExternalEndpoint string
	AccessKey        string
	SecretKey        string
	UseSSL           bool
}

// NewMinIOStorage creates a new MinIO storage client.
func NewMinIOStorage(cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinIOStorage{
		client:           client,
		externalEndpoint: cfg.ExternalEndpoint,
		useSSL:           cfg.UseSSL,
	}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (s *MinIOStorage) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}

func (s *MinIOStorage) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinIOStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
}

func (s *MinIOStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	return s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) PresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	u, err := s.client.PresignedPutObject(ctx, bucket, key, expires)
	if err != nil {
		return "", err
	}
	return s.rewriteURL(u), nil
}

func (s *MinIOStorage) PresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	u, err := s.client.PresignedGetObject(ctx, bucket, key, expires, reqParams)
	if err != nil {
		return "", err
	}
	return s.rewriteURL(u), nil
}

// rewriteURL replaces the internal MinIO host with the external endpoint
// so that presigned URLs are accessible from the browser.
func (s *MinIOStorage) rewriteURL(u *url.URL) string {
	if s.externalEndpoint == "" {
		return u.String()
	}
	u.Host = s.externalEndpoint
	if s.useSSL {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}
	return u.String()
}

func (s *MinIOStorage) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *MinIOStorage) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
	}, nil
}
