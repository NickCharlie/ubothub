package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PendingOrderRepository defines the data access interface for pending orders.
type PendingOrderRepository interface {
	Create(ctx context.Context, order *model.PendingOrder) error
	FindByID(ctx context.Context, id string) (*model.PendingOrder, error)
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*model.PendingOrder, error)
	UpdateStatus(ctx context.Context, tx *gorm.DB, id, status string) error
	SetCompleted(ctx context.Context, tx *gorm.DB, id, externalID string) error
	FindPendingByUserID(ctx context.Context, userID string, page, pageSize int) ([]model.PendingOrder, int64, error)
	FindExpired(ctx context.Context, limit int) ([]model.PendingOrder, error)
}

type pendingOrderRepository struct {
	db *gorm.DB
}

// NewPendingOrderRepository creates a new pending order repository.
func NewPendingOrderRepository(db *gorm.DB) PendingOrderRepository {
	return &pendingOrderRepository{db: db}
}

func (r *pendingOrderRepository) Create(ctx context.Context, order *model.PendingOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *pendingOrderRepository) FindByID(ctx context.Context, id string) (*model.PendingOrder, error) {
	var order model.PendingOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *pendingOrderRepository) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*model.PendingOrder, error) {
	var order model.PendingOrder
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *pendingOrderRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, id, status string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).
		Model(&model.PendingOrder{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *pendingOrderRepository) SetCompleted(ctx context.Context, tx *gorm.DB, id, externalID string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	now := gorm.Expr("NOW()")
	return db.WithContext(ctx).
		Model(&model.PendingOrder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.OrderStatusPaid,
			"external_id":  externalID,
			"completed_at": now,
		}).Error
}

func (r *pendingOrderRepository) FindPendingByUserID(ctx context.Context, userID string, page, pageSize int) ([]model.PendingOrder, int64, error) {
	var orders []model.PendingOrder
	var total int64

	query := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.OrderStatusPending)

	if err := query.Model(&model.PendingOrder{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error
	return orders, total, err
}

func (r *pendingOrderRepository) FindExpired(ctx context.Context, limit int) ([]model.PendingOrder, error) {
	var orders []model.PendingOrder
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < NOW()", model.OrderStatusPending).
		Limit(limit).
		Find(&orders).Error
	return orders, err
}
