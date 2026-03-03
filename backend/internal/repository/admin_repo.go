package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// AdminRepository defines the data access interface for admin operations.
type AdminRepository interface {
	CountUsers(ctx context.Context) (int64, error)
	CountBots(ctx context.Context) (int64, error)
	CountActiveBots(ctx context.Context) (int64, error)
	CountAssets(ctx context.Context) (int64, error)
	SumAssetFileSize(ctx context.Context) (int64, error)

	ListUsers(ctx context.Context, offset, limit int, status, role string) ([]*model.User, int64, error)
	UpdateUserStatus(ctx context.Context, userID, status string) error

	ListBots(ctx context.Context, offset, limit int, status, framework string) ([]*model.Bot, int64, error)
	ForceDeleteBot(ctx context.Context, botID string) error

	ListAssets(ctx context.Context, offset, limit int, category, status string) ([]*model.Asset, int64, error)

	ListMessageLogs(ctx context.Context, offset, limit int, botID string) ([]*model.MessageLog, int64, error)
}

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new GORM-backed admin repository.
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) CountBots(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Bot{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) CountActiveBots(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Bot{}).Where("status = ?", "online").Count(&count).Error
	return count, err
}

func (r *adminRepository) CountAssets(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Asset{}).Count(&count).Error
	return count, err
}

func (r *adminRepository) SumAssetFileSize(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Asset{}).Select("COALESCE(SUM(file_size), 0)").Scan(&total).Error
	return total, err
}

func (r *adminRepository) ListUsers(ctx context.Context, offset, limit int, status, role string) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	db := r.db.WithContext(ctx).Model(&model.User{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if role != "" {
		db = db.Where("role = ?", role)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *adminRepository) UpdateUserStatus(ctx context.Context, userID, status string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("status", status).Error
}

func (r *adminRepository) ListBots(ctx context.Context, offset, limit int, status, framework string) ([]*model.Bot, int64, error) {
	var bots []*model.Bot
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Bot{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if framework != "" {
		db = db.Where("framework = ?", framework)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&bots).Error; err != nil {
		return nil, 0, err
	}
	return bots, total, nil
}

func (r *adminRepository) ForceDeleteBot(ctx context.Context, botID string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.Bot{}, "id = ?", botID).Error
}

func (r *adminRepository) ListAssets(ctx context.Context, offset, limit int, category, status string) ([]*model.Asset, int64, error) {
	var assets []*model.Asset
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Asset{})
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *adminRepository) ListMessageLogs(ctx context.Context, offset, limit int, botID string) ([]*model.MessageLog, int64, error) {
	var logs []*model.MessageLog
	var total int64

	db := r.db.WithContext(ctx).Model(&model.MessageLog{})
	if botID != "" {
		db = db.Where("bot_id = ?", botID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
