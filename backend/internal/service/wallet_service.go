package service

import (
	"context"
	"fmt"

	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WalletService handles wallet operations with transactional consistency.
type WalletService struct {
	walletRepo repository.WalletRepository
	txnRepo    repository.TransactionRepository
	db         *gorm.DB
	logger     *zap.Logger
}

// NewWalletService creates a new wallet service.
func NewWalletService(
	walletRepo repository.WalletRepository,
	txnRepo repository.TransactionRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
		txnRepo:    txnRepo,
		db:         db,
		logger:     logger,
	}
}

// GetOrCreateWallet returns the user's wallet, creating one if it does not exist.
func (s *WalletService) GetOrCreateWallet(ctx context.Context, userID string) (*model.Wallet, error) {
	wallet, err := s.walletRepo.FindByUserID(ctx, userID)
	if err == nil {
		return wallet, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find wallet: %w", err)
	}

	wallet = &model.Wallet{
		ID:       xid.New().String(),
		UserID:   userID,
		Balance:  decimal.Zero,
		Frozen:   decimal.Zero,
		Currency: "CNY",
	}
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}

	s.logger.Info("wallet created", zap.String("user_id", userID), zap.String("wallet_id", wallet.ID))
	return wallet, nil
}

// TopUp adds funds to the user's wallet after payment confirmation.
// This method runs within a database transaction to ensure consistency.
func (s *WalletService) TopUp(ctx context.Context, userID string, amount decimal.Decimal, externalOrderID, channel, description string) (*model.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	var txn *model.Transaction
	err := s.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		wallet, err := s.walletRepo.FindByUserIDForUpdate(ctx, dbTx, userID)
		if err != nil {
			return fmt.Errorf("lock wallet: %w", err)
		}

		balanceBefore := wallet.Balance
		balanceAfter := wallet.Balance.Add(amount)

		txn = &model.Transaction{
			ID:              xid.New().String(),
			UserID:          userID,
			WalletID:        wallet.ID,
			Type:            model.TxTypeTopUp,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceAfter,
			Status:          model.TxStatusCompleted,
			ExternalOrderID: externalOrderID,
			PaymentChannel:  channel,
			Description:     description,
		}

		if err := s.txnRepo.Create(ctx, dbTx, txn); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(ctx, dbTx, wallet.ID, balanceAfter, wallet.Frozen); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("wallet topped up",
		zap.String("user_id", userID),
		zap.String("amount", amount.String()),
		zap.String("channel", channel),
	)
	return txn, nil
}

// Deduct subtracts funds from the user's wallet for bot usage.
func (s *WalletService) Deduct(ctx context.Context, userID string, amount decimal.Decimal, botID, description string) (*model.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	var txn *model.Transaction
	err := s.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		wallet, err := s.walletRepo.FindByUserIDForUpdate(ctx, dbTx, userID)
		if err != nil {
			return fmt.Errorf("lock wallet: %w", err)
		}

		if wallet.AvailableBalance().LessThan(amount) {
			return ErrInsufficientBalance
		}

		balanceBefore := wallet.Balance
		balanceAfter := wallet.Balance.Sub(amount)

		txn = &model.Transaction{
			ID:            xid.New().String(),
			UserID:        userID,
			WalletID:      wallet.ID,
			Type:          model.TxTypeUsage,
			Amount:        amount.Neg(),
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Status:        model.TxStatusCompleted,
			BotID:         botID,
			Description:   description,
		}

		if err := s.txnRepo.Create(ctx, dbTx, txn); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(ctx, dbTx, wallet.ID, balanceAfter, wallet.Frozen); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Debug("wallet deducted",
		zap.String("user_id", userID),
		zap.String("amount", amount.String()),
		zap.String("bot_id", botID),
	)
	return txn, nil
}

// Credit adds earnings to the creator's wallet from bot usage revenue.
func (s *WalletService) Credit(ctx context.Context, userID string, amount decimal.Decimal, botID, description string) (*model.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	var txn *model.Transaction
	err := s.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		wallet, err := s.walletRepo.FindByUserIDForUpdate(ctx, dbTx, userID)
		if err != nil {
			return fmt.Errorf("lock wallet: %w", err)
		}

		balanceBefore := wallet.Balance
		balanceAfter := wallet.Balance.Add(amount)

		txn = &model.Transaction{
			ID:            xid.New().String(),
			UserID:        userID,
			WalletID:      wallet.ID,
			Type:          model.TxTypeEarning,
			Amount:        amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Status:        model.TxStatusCompleted,
			BotID:         botID,
			Description:   description,
		}

		if err := s.txnRepo.Create(ctx, dbTx, txn); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(ctx, dbTx, wallet.ID, balanceAfter, wallet.Frozen); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("creator earnings credited",
		zap.String("user_id", userID),
		zap.String("amount", amount.String()),
		zap.String("bot_id", botID),
	)
	return txn, nil
}

// GetTransactions returns paginated transaction history for a user.
func (s *WalletService) GetTransactions(ctx context.Context, userID string, page, pageSize int, txType string) ([]*model.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.txnRepo.FindByUserID(ctx, userID, offset, pageSize, txType)
}
