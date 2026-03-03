package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// BotPricingRepository defines the data access interface for BotPricing entities.
type BotPricingRepository interface {
	Create(ctx context.Context, pricing *model.BotPricing) error
	FindByBotID(ctx context.Context, botID string) (*model.BotPricing, error)
	Update(ctx context.Context, pricing *model.BotPricing) error
	Delete(ctx context.Context, botID string) error
}

type botPricingRepository struct {
	db *gorm.DB
}

// NewBotPricingRepository creates a new GORM-backed bot pricing repository.
func NewBotPricingRepository(db *gorm.DB) BotPricingRepository {
	return &botPricingRepository{db: db}
}

func (r *botPricingRepository) Create(ctx context.Context, pricing *model.BotPricing) error {
	return r.db.WithContext(ctx).Create(pricing).Error
}

func (r *botPricingRepository) FindByBotID(ctx context.Context, botID string) (*model.BotPricing, error) {
	var pricing model.BotPricing
	err := r.db.WithContext(ctx).Where("bot_id = ?", botID).First(&pricing).Error
	if err != nil {
		return nil, err
	}
	return &pricing, nil
}

func (r *botPricingRepository) Update(ctx context.Context, pricing *model.BotPricing) error {
	return r.db.WithContext(ctx).Save(pricing).Error
}

func (r *botPricingRepository) Delete(ctx context.Context, botID string) error {
	return r.db.WithContext(ctx).Where("bot_id = ?", botID).Delete(&model.BotPricing{}).Error
}
