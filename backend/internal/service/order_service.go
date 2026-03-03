package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderService manages pending payment orders lifecycle.
type OrderService struct {
	orderRepo  repository.PendingOrderRepository
	walletRepo repository.WalletRepository
	walletSvc  *WalletService
	db         *gorm.DB
	logger     *zap.Logger
}

// NewOrderService creates a new order service.
func NewOrderService(
	orderRepo repository.PendingOrderRepository,
	walletRepo repository.WalletRepository,
	walletSvc *WalletService,
	db *gorm.DB,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		orderRepo:  orderRepo,
		walletRepo: walletRepo,
		walletSvc:  walletSvc,
		db:         db,
		logger:     logger,
	}
}

// CreateOrder creates a pending order before sending to payment provider.
func (s *OrderService) CreateOrder(ctx context.Context, userID string, amount decimal.Decimal, channel, description string) (*model.PendingOrder, error) {
	wallet, err := s.walletSvc.GetOrCreateWallet(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	order := &model.PendingOrder{
		ID:             xid.New().String(),
		UserID:         userID,
		WalletID:       wallet.ID,
		Amount:         amount,
		Status:         model.OrderStatusPending,
		PaymentChannel: channel,
		Description:    description,
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	s.logger.Info("pending order created",
		zap.String("order_id", order.ID),
		zap.String("user_id", userID),
		zap.String("amount", amount.String()),
		zap.String("channel", channel),
	)
	return order, nil
}

// CompleteOrder processes a confirmed payment: marks the order as paid and tops up the wallet.
// This method is idempotent — if the order is already paid, it returns nil without double-crediting.
func (s *OrderService) CompleteOrder(ctx context.Context, orderID, externalID string, confirmedAmount decimal.Decimal) error {
	return s.db.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		order, err := s.orderRepo.FindByIDForUpdate(ctx, dbTx, orderID)
		if err != nil {
			return fmt.Errorf("find order: %w", err)
		}

		// Idempotency: already completed.
		if order.Status == model.OrderStatusPaid {
			s.logger.Debug("order already completed, skipping",
				zap.String("order_id", orderID),
			)
			return nil
		}

		if order.Status != model.OrderStatusPending {
			return fmt.Errorf("order %s has status %s, cannot complete", orderID, order.Status)
		}

		// Mark order as completed.
		if err := s.orderRepo.SetCompleted(ctx, dbTx, orderID, externalID); err != nil {
			return fmt.Errorf("set completed: %w", err)
		}

		// Top up the user's wallet.
		_, err = s.walletSvc.TopUp(ctx, order.UserID, confirmedAmount, orderID, order.PaymentChannel, order.Description)
		if err != nil {
			return fmt.Errorf("top up wallet: %w", err)
		}

		s.logger.Info("payment completed, wallet topped up",
			zap.String("order_id", orderID),
			zap.String("user_id", order.UserID),
			zap.String("amount", confirmedAmount.String()),
			zap.String("external_id", externalID),
		)
		return nil
	})
}

// FailOrder marks an order as failed.
func (s *OrderService) FailOrder(ctx context.Context, orderID string) error {
	return s.orderRepo.UpdateStatus(ctx, nil, orderID, model.OrderStatusFailed)
}

// ExpireOrders marks expired pending orders.
func (s *OrderService) ExpireOrders(ctx context.Context) (int, error) {
	orders, err := s.orderRepo.FindExpired(ctx, 100)
	if err != nil {
		return 0, fmt.Errorf("find expired: %w", err)
	}

	count := 0
	for _, order := range orders {
		if err := s.orderRepo.UpdateStatus(ctx, nil, order.ID, model.OrderStatusExpired); err != nil {
			s.logger.Warn("failed to expire order",
				zap.String("order_id", order.ID),
				zap.Error(err),
			)
			continue
		}
		count++
	}

	if count > 0 {
		s.logger.Info("expired pending orders", zap.Int("count", count))
	}
	return count, nil
}
