package repository

import (
	"context"

	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
)

// TransactionRepository defines the data access interface for Transaction entities.
type TransactionRepository interface {
	Create(ctx context.Context, tx *gorm.DB, txn *model.Transaction) error
	FindByID(ctx context.Context, id string) (*model.Transaction, error)
	FindByExternalOrderID(ctx context.Context, externalID string) (*model.Transaction, error)
	UpdateStatus(ctx context.Context, tx *gorm.DB, id, status string) error
	FindByUserID(ctx context.Context, userID string, offset, limit int, txType string) ([]*model.Transaction, int64, error)
}

type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new GORM-backed transaction repository.
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *gorm.DB, txn *model.Transaction) error {
	if tx != nil {
		return tx.WithContext(ctx).Create(txn).Error
	}
	return r.db.WithContext(ctx).Create(txn).Error
}

func (r *transactionRepository) FindByID(ctx context.Context, id string) (*model.Transaction, error) {
	var txn model.Transaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&txn).Error
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *transactionRepository) FindByExternalOrderID(ctx context.Context, externalID string) (*model.Transaction, error) {
	var txn model.Transaction
	err := r.db.WithContext(ctx).Where("external_order_id = ?", externalID).First(&txn).Error
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, id, status string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Model(&model.Transaction{}).Where("id = ?", id).Update("status", status).Error
}

func (r *transactionRepository) FindByUserID(ctx context.Context, userID string, offset, limit int, txType string) ([]*model.Transaction, int64, error) {
	var txns []*model.Transaction
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Transaction{}).Where("user_id = ?", userID)
	if txType != "" {
		db = db.Where("type = ?", txType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&txns).Error; err != nil {
		return nil, 0, err
	}
	return txns, total, nil
}
