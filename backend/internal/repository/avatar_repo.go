package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// AvatarRepository defines the data access interface for AvatarConfig entities.
type AvatarRepository interface {
	Create(ctx context.Context, avatar *model.AvatarConfig) error
	FindByID(ctx context.Context, id string) (*model.AvatarConfig, error)
	FindByIDWithAssets(ctx context.Context, id string) (*model.AvatarConfig, error)
	FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*model.AvatarConfig, int64, error)
	FindByBotID(ctx context.Context, botID string) (*model.AvatarConfig, error)
	Update(ctx context.Context, avatar *model.AvatarConfig) error
	Delete(ctx context.Context, id string) error
	CreateAvatarAsset(ctx context.Context, aa *model.AvatarAsset) error
	DeleteAvatarAsset(ctx context.Context, avatarID, assetID string) error
	FindAvatarAssets(ctx context.Context, avatarID string) ([]model.AvatarAsset, error)
}

type avatarRepository struct {
	db *gorm.DB
}

// NewAvatarRepository creates a new GORM-backed avatar repository.
func NewAvatarRepository(db *gorm.DB) AvatarRepository {
	return &avatarRepository{db: db}
}

func (r *avatarRepository) Create(ctx context.Context, avatar *model.AvatarConfig) error {
	return r.db.WithContext(ctx).Create(avatar).Error
}

func (r *avatarRepository) FindByID(ctx context.Context, id string) (*model.AvatarConfig, error) {
	var avatar model.AvatarConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *avatarRepository) FindByIDWithAssets(ctx context.Context, id string) (*model.AvatarConfig, error) {
	var avatar model.AvatarConfig
	err := r.db.WithContext(ctx).
		Preload("AvatarAssets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("AvatarAssets.Asset").
		Where("id = ?", id).
		First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *avatarRepository) FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*model.AvatarConfig, int64, error) {
	var avatars []*model.AvatarConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AvatarConfig{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&avatars).Error; err != nil {
		return nil, 0, err
	}
	return avatars, total, nil
}

func (r *avatarRepository) FindByBotID(ctx context.Context, botID string) (*model.AvatarConfig, error) {
	var avatar model.AvatarConfig
	err := r.db.WithContext(ctx).Where("bot_id = ?", botID).First(&avatar).Error
	if err != nil {
		return nil, err
	}
	return &avatar, nil
}

func (r *avatarRepository) Update(ctx context.Context, avatar *model.AvatarConfig) error {
	return r.db.WithContext(ctx).Save(avatar).Error
}

func (r *avatarRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AvatarConfig{}).Error
}

func (r *avatarRepository) CreateAvatarAsset(ctx context.Context, aa *model.AvatarAsset) error {
	return r.db.WithContext(ctx).Create(aa).Error
}

func (r *avatarRepository) DeleteAvatarAsset(ctx context.Context, avatarID, assetID string) error {
	return r.db.WithContext(ctx).
		Where("avatar_id = ? AND asset_id = ?", avatarID, assetID).
		Delete(&model.AvatarAsset{}).Error
}

func (r *avatarRepository) FindAvatarAssets(ctx context.Context, avatarID string) ([]model.AvatarAsset, error) {
	var assets []model.AvatarAsset
	err := r.db.WithContext(ctx).
		Preload("Asset").
		Where("avatar_id = ?", avatarID).
		Order("sort_order ASC").
		Find(&assets).Error
	return assets, err
}
