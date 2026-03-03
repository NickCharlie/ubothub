package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliyunOSSStorage implements ObjectStorage using Alibaba Cloud OSS.
type AliyunOSSStorage struct {
	client    *oss.Client
	cdnDomain string
}

// AliyunOSSConfig holds connection parameters for Alibaba Cloud OSS.
type AliyunOSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	CDNDomain       string
}

// NewAliyunOSSStorage creates a new Alibaba Cloud OSS storage client.
func NewAliyunOSSStorage(cfg AliyunOSSConfig) (*AliyunOSSStorage, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &AliyunOSSStorage{
		client:    client,
		cdnDomain: cfg.CDNDomain,
	}, nil
}

func (s *AliyunOSSStorage) getBucket(bucket string) (*oss.Bucket, error) {
	return s.client.Bucket(bucket)
}

func (s *AliyunOSSStorage) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}
	return b.PutObject(key, reader, oss.ContentType(contentType))
}

func (s *AliyunOSSStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	b, err := s.getBucket(bucket)
	if err != nil {
		return nil, err
	}
	return b.GetObject(key)
}

func (s *AliyunOSSStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}
	return b.DeleteObject(key)
}

func (s *AliyunOSSStorage) PresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	b, err := s.getBucket(bucket)
	if err != nil {
		return "", err
	}
	url, err := b.SignURL(key, oss.HTTPPut, int64(expires.Seconds()))
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *AliyunOSSStorage) PresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	b, err := s.getBucket(bucket)
	if err != nil {
		return "", err
	}
	url, err := b.SignURL(key, oss.HTTPGet, int64(expires.Seconds()))
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *AliyunOSSStorage) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	b, err := s.getBucket(bucket)
	if err != nil {
		return false, err
	}
	return b.IsObjectExist(key)
}

func (s *AliyunOSSStorage) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	b, err := s.getBucket(bucket)
	if err != nil {
		return nil, err
	}

	header, err := b.GetObjectMeta(key)
	if err != nil {
		return nil, err
	}

	var size int64
	fmt.Sscanf(header.Get("Content-Length"), "%d", &size)

	lastModified, _ := time.Parse(time.RFC1123, header.Get("Last-Modified"))

	return &ObjectInfo{
		Key:          key,
		Size:         size,
		ContentType:  header.Get("Content-Type"),
		LastModified: lastModified,
	}, nil
}

// Compile-time interface compliance checks.
var _ ObjectStorage = (*AliyunOSSStorage)(nil)
var _ ObjectStorage = (*MinIOStorage)(nil)
