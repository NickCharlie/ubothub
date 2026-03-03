package storage

import (
	"fmt"

	"github.com/ubothub/backend/internal/config"
)

// NewFromConfig creates an ObjectStorage implementation based on the
// configured provider. Supported providers: "minio", "aliyun_oss".
func NewFromConfig(cfg config.StorageConfig) (ObjectStorage, error) {
	switch cfg.Provider {
	case "minio":
		return NewMinIOStorage(MinIOConfig{
			Endpoint:  cfg.MinIO.Endpoint,
			AccessKey: cfg.MinIO.AccessKey,
			SecretKey: cfg.MinIO.SecretKey,
			UseSSL:    cfg.MinIO.UseSSL,
		})
	case "aliyun_oss":
		return NewAliyunOSSStorage(AliyunOSSConfig{
			Endpoint:        cfg.AliyunOSS.Endpoint,
			AccessKeyID:     cfg.AliyunOSS.AccessKeyID,
			AccessKeySecret: cfg.AliyunOSS.AccessKeySecret,
			CDNDomain:       cfg.AliyunOSS.CDNDomain,
		})
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", cfg.Provider)
	}
}
