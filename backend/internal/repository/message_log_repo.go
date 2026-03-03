package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// MessageLogRepository defines the data access interface for MessageLog entities.
type MessageLogRepository interface {
	Create(ctx context.Context, log *model.MessageLog) error
	FindByBotID(ctx context.Context, botID string, offset, limit int) ([]*model.MessageLog, int64, error)
}

type messageLogRepository struct {
	db *gorm.DB
}

// NewMessageLogRepository creates a new GORM-backed message log repository.
func NewMessageLogRepository(db *gorm.DB) MessageLogRepository {
	return &messageLogRepository{db: db}
}

func (r *messageLogRepository) Create(ctx context.Context, log *model.MessageLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *messageLogRepository) FindByBotID(ctx context.Context, botID string, offset, limit int) ([]*model.MessageLog, int64, error) {
	var logs []*model.MessageLog
	var total int64

	db := r.db.WithContext(ctx).Model(&model.MessageLog{}).Where("bot_id = ?", botID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
