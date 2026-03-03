package repository

import (
	"context"

	"github.com/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// AssetRepository defines the data access interface for Asset entities.
type AssetRepository interface {
	Create(ctx context.Context, asset *model.Asset) error
	FindByID(ctx context.Context, id string) (*model.Asset, error)
	FindByUserID(ctx context.Context, userID string, offset, limit int, category, format, status string) ([]*model.Asset, int64, error)
	FindPublic(ctx context.Context, offset, limit int, category, format, search string) ([]*model.Asset, int64, error)
	Update(ctx context.Context, asset *model.Asset) error
	Delete(ctx context.Context, id string) error
	SumFileSizeByUserID(ctx context.Context, userID string) (int64, error)
}

type assetRepository struct {
	db *gorm.DB
}

// NewAssetRepository creates a new GORM-backed asset repository.
func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{db: db}
}

func (r *assetRepository) Create(ctx context.Context, asset *model.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *assetRepository) FindByID(ctx context.Context, id string) (*model.Asset, error) {
	var asset model.Asset
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *assetRepository) FindByUserID(ctx context.Context, userID string, offset, limit int, category, format, status string) ([]*model.Asset, int64, error) {
	var assets []*model.Asset
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Asset{}).Where("user_id = ?", userID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if format != "" {
		query = query.Where("format = ?", format)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *assetRepository) FindPublic(ctx context.Context, offset, limit int, category, format, search string) ([]*model.Asset, int64, error) {
	var assets []*model.Asset
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Asset{}).Where("is_public = ? AND status = ?", true, "ready")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if format != "" {
		query = query.Where("format = ?", format)
	}
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("download_count DESC, created_at DESC").Offset(offset).Limit(limit).Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *assetRepository) Update(ctx context.Context, asset *model.Asset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

func (r *assetRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Asset{}).Error
}

func (r *assetRepository) SumFileSizeByUserID(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Asset{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&total).Error
	return total, err
}
