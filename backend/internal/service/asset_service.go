package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/internal/storage"
	"github.com/NickCharlie/ubothub/backend/pkg/sanitize"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	maxFileSizeBytes   = 500 * 1024 * 1024 // 500 MB per file
	maxQuotaBytes      = 5 * 1024 * 1024 * 1024 // 5 GB per user
	presignedURLExpiry = 1 * time.Hour
)

// Allowed file extensions per category (whitelist).
var allowedFormats = map[string]map[string]bool{
	"model_3d": {
		"vrm":  true,
		"glb":  true,
		"gltf": true,
		"fbx":  true,
	},
	"model_live2d": {
		"zip":  true, // Live2D packaged as ZIP (moc3 + model3.json + textures)
		"moc3": true,
	},
	"motion": {
		"bvh":  true,
		"vmd":  true,
		"fbx":  true,
		"vrma": true,
	},
	"texture": {
		"png":  true,
		"jpg":  true,
		"jpeg": true,
		"webp": true,
		"ktx2": true,
	},
}

// AssetService handles asset management business logic.
type AssetService struct {
	assetRepo repository.AssetRepository
	storage   storage.ObjectStorage
	bucket    string
	logger    *zap.Logger
}

// NewAssetService creates a new asset service.
func NewAssetService(
	assetRepo repository.AssetRepository,
	store storage.ObjectStorage,
	bucket string,
	logger *zap.Logger,
) *AssetService {
	return &AssetService{
		assetRepo: assetRepo,
		storage:   store,
		bucket:    bucket,
		logger:    logger,
	}
}

// ValidateUpload checks file size, format, and user quota before upload.
func (s *AssetService) ValidateUpload(ctx context.Context, userID, filename, category string, fileSize int64) error {
	if fileSize > maxFileSizeBytes {
		return ErrAssetSizeTooLarge
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	formats, ok := allowedFormats[category]
	if !ok {
		return ErrAssetFormatInvalid
	}
	if !formats[ext] {
		return ErrAssetFormatInvalid
	}

	usedBytes, err := s.assetRepo.SumFileSizeByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("check quota: %w", err)
	}
	if usedBytes+fileSize > maxQuotaBytes {
		return ErrAssetQuotaExceeded
	}

	return nil
}

// GeneratePresignedUpload creates a presigned PUT URL for client-side direct upload.
func (s *AssetService) GeneratePresignedUpload(ctx context.Context, userID, filename, category string, fileSize int64) (string, string, error) {
	if err := s.ValidateUpload(ctx, userID, filename, category, fileSize); err != nil {
		return "", "", err
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	assetID := xid.New().String()
	fileKey := fmt.Sprintf("%s/%s/%s.%s", userID, assetID, assetID, ext)

	url, err := s.storage.PresignedPutURL(ctx, s.bucket, fileKey, presignedURLExpiry)
	if err != nil {
		return "", "", fmt.Errorf("generate presigned URL: %w", err)
	}

	s.logger.Debug("presigned upload URL generated",
		zap.String("user_id", userID),
		zap.String("file_key", fileKey),
	)

	return url, fileKey, nil
}

// CompleteUpload creates the asset record after the client confirms upload.
func (s *AssetService) CompleteUpload(ctx context.Context, userID, fileKey, name, description, category, format string, fileSize int64, isPublic bool, tags []string) (*model.Asset, error) {
	ext := strings.ToLower(format)
	formats, ok := allowedFormats[category]
	if !ok || !formats[ext] {
		return nil, ErrAssetFormatInvalid
	}

	exists, err := s.storage.ObjectExists(ctx, s.bucket, fileKey)
	if err != nil {
		return nil, fmt.Errorf("check object existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("uploaded file not found at key: %s", fileKey)
	}

	asset := &model.Asset{
		ID:          xid.New().String(),
		UserID:      userID,
		Name:        sanitize.Text(name),
		Description: sanitize.Text(description),
		Category:    category,
		Format:      ext,
		FileKey:     fileKey,
		FileSize:    fileSize,
		Metadata:    "{}",
		Tags:        tags,
		IsPublic:    isPublic,
		Status:      "processing",
	}

	if err := s.assetRepo.Create(ctx, asset); err != nil {
		return nil, fmt.Errorf("create asset: %w", err)
	}

	s.logger.Info("asset upload completed",
		zap.String("asset_id", asset.ID),
		zap.String("user_id", userID),
		zap.String("category", category),
		zap.String("format", ext),
	)

	return asset, nil
}

// GetAsset returns an asset by ID, verifying ownership for non-public assets.
func (s *AssetService) GetAsset(ctx context.Context, assetID, userID string) (*model.Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("find asset: %w", err)
	}

	if !asset.IsPublic && asset.UserID != userID {
		return nil, ErrAssetNotFound
	}

	return asset, nil
}

// ListAssets returns paginated assets for a user.
func (s *AssetService) ListAssets(ctx context.Context, userID string, page, pageSize int, category, format, status string) ([]*model.Asset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.assetRepo.FindByUserID(ctx, userID, offset, pageSize, category, format, status)
}

// ListPublicAssets returns paginated public assets.
func (s *AssetService) ListPublicAssets(ctx context.Context, page, pageSize int, category, format, search string) ([]*model.Asset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.assetRepo.FindPublic(ctx, offset, pageSize, category, format, search)
}

// UpdateAsset updates asset metadata.
func (s *AssetService) UpdateAsset(ctx context.Context, assetID, userID, name, description string, isPublic *bool, tags []string) (*model.Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAssetNotFound
		}
		return nil, fmt.Errorf("find asset: %w", err)
	}

	if asset.UserID != userID {
		return nil, ErrAssetNotFound
	}

	if name != "" {
		asset.Name = sanitize.Text(name)
	}
	if description != "" {
		asset.Description = sanitize.Text(description)
	}
	if isPublic != nil {
		asset.IsPublic = *isPublic
	}
	if tags != nil {
		asset.Tags = tags
	}

	if err := s.assetRepo.Update(ctx, asset); err != nil {
		return nil, fmt.Errorf("update asset: %w", err)
	}

	return asset, nil
}

// DeleteAsset deletes an asset and its stored file.
func (s *AssetService) DeleteAsset(ctx context.Context, assetID, userID string) error {
	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAssetNotFound
		}
		return fmt.Errorf("find asset: %w", err)
	}

	if asset.UserID != userID {
		return ErrAssetNotFound
	}

	if err := s.storage.DeleteObject(ctx, s.bucket, asset.FileKey); err != nil {
		s.logger.Warn("failed to delete stored object",
			zap.String("asset_id", assetID),
			zap.String("file_key", asset.FileKey),
			zap.Error(err),
		)
	}

	if asset.ThumbnailKey != "" {
		_ = s.storage.DeleteObject(ctx, s.bucket, asset.ThumbnailKey)
	}

	if err := s.assetRepo.Delete(ctx, assetID); err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}

	s.logger.Info("asset deleted", zap.String("asset_id", assetID), zap.String("user_id", userID))
	return nil
}

// GetDownloadURL returns a presigned download URL for the asset.
func (s *AssetService) GetDownloadURL(ctx context.Context, assetID, userID string) (string, error) {
	asset, err := s.GetAsset(ctx, assetID, userID)
	if err != nil {
		return "", err
	}

	url, err := s.storage.PresignedGetURL(ctx, s.bucket, asset.FileKey, presignedURLExpiry)
	if err != nil {
		return "", fmt.Errorf("generate download URL: %w", err)
	}

	return url, nil
}

// GetThumbnailURL returns a presigned thumbnail URL.
func (s *AssetService) GetThumbnailURL(ctx context.Context, assetID, userID string) (string, error) {
	asset, err := s.GetAsset(ctx, assetID, userID)
	if err != nil {
		return "", err
	}

	if asset.ThumbnailKey == "" {
		return "", ErrAssetNotFound
	}

	url, err := s.storage.PresignedGetURL(ctx, s.bucket, asset.ThumbnailKey, presignedURLExpiry)
	if err != nil {
		return "", fmt.Errorf("generate thumbnail URL: %w", err)
	}

	return url, nil
}
