package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// BotRepository defines the data access interface for Bot entities.
type BotRepository interface {
	Create(ctx context.Context, bot *model.Bot) error
	FindByID(ctx context.Context, id string) (*model.Bot, error)
	FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*model.Bot, int64, error)
	FindByAccessToken(ctx context.Context, token string) (*model.Bot, error)
	Update(ctx context.Context, bot *model.Bot) error
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

type botRepository struct {
	db *gorm.DB
}

// NewBotRepository creates a new GORM-backed bot repository.
func NewBotRepository(db *gorm.DB) BotRepository {
	return &botRepository{db: db}
}

func (r *botRepository) Create(ctx context.Context, bot *model.Bot) error {
	return r.db.WithContext(ctx).Create(bot).Error
}

func (r *botRepository) FindByID(ctx context.Context, id string) (*model.Bot, error) {
	var bot model.Bot
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) FindByUserID(ctx context.Context, userID string, offset, limit int) ([]*model.Bot, int64, error) {
	var bots []*model.Bot
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Bot{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&bots).Error; err != nil {
		return nil, 0, err
	}
	return bots, total, nil
}

func (r *botRepository) FindByAccessToken(ctx context.Context, token string) (*model.Bot, error) {
	var bot model.Bot
	err := r.db.WithContext(ctx).Where("access_token = ?", token).First(&bot).Error
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *botRepository) Update(ctx context.Context, bot *model.Bot) error {
	return r.db.WithContext(ctx).Save(bot).Error
}

func (r *botRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Bot{}).Error
}

func (r *botRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Bot{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
