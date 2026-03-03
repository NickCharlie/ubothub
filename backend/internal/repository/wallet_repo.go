package repository

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalletRepository defines the data access interface for Wallet entities.
type WalletRepository interface {
	Create(ctx context.Context, wallet *model.Wallet) error
	FindByUserID(ctx context.Context, userID string) (*model.Wallet, error)
	FindByUserIDForUpdate(ctx context.Context, tx *gorm.DB, userID string) (*model.Wallet, error)
	UpdateBalance(ctx context.Context, tx *gorm.DB, walletID string, balance, frozen decimal.Decimal) error
}

type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new GORM-backed wallet repository.
func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *model.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *walletRepository) FindByUserID(ctx context.Context, userID string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// FindByUserIDForUpdate acquires a row-level lock for concurrent balance updates.
func (r *walletRepository) FindByUserIDForUpdate(ctx context.Context, tx *gorm.DB, userID string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, tx *gorm.DB, walletID string, balance, frozen decimal.Decimal) error {
	result := tx.WithContext(ctx).Model(&model.Wallet{}).
		Where("id = ?", walletID).
		Updates(map[string]interface{}{
			"balance": balance,
			"frozen":  frozen,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("wallet not found: %s", walletID)
	}
	return nil
}
