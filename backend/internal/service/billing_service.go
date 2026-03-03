package service

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
)

// BillingService handles bot usage billing with revenue sharing between
// the platform and bot creators.
type BillingService struct {
	pricingRepo repository.BotPricingRepository
	walletSvc   *WalletService
	botRepo     repository.BotRepository
	logger      *zap.Logger
}

// NewBillingService creates a new billing service.
func NewBillingService(
	pricingRepo repository.BotPricingRepository,
	walletSvc *WalletService,
	botRepo repository.BotRepository,
	logger *zap.Logger,
) *BillingService {
	return &BillingService{
		pricingRepo: pricingRepo,
		walletSvc:   walletSvc,
		botRepo:     botRepo,
		logger:      logger,
	}
}

// BillResult holds the result of a billing operation.
type BillResult struct {
	Charged      bool            `json:"charged"`
	Amount       decimal.Decimal `json:"amount"`
	CreatorShare decimal.Decimal `json:"creator_share"`
	PlatformFee  decimal.Decimal `json:"platform_fee"`
}

// ChargeForCall processes billing for a single bot call.
// Returns nil error and Charged=false for free bots or free call quotas.
func (s *BillingService) ChargeForCall(ctx context.Context, callerUserID, botID string) (*BillResult, error) {
	pricing, err := s.pricingRepo.FindByBotID(ctx, botID)
	if err != nil {
		// No pricing configured means the bot is free.
		return &BillResult{Charged: false}, nil
	}

	if pricing.Mode == model.PricingModeFree {
		return &BillResult{Charged: false}, nil
	}

	if pricing.Mode != model.PricingModePerCall {
		return &BillResult{Charged: false}, nil
	}

	if pricing.PricePerCall.LessThanOrEqual(decimal.Zero) {
		return &BillResult{Charged: false}, nil
	}

	// Retrieve the bot to find the owner (creator).
	bot, err := s.botRepo.FindByID(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("find bot: %w", err)
	}

	amount := pricing.PricePerCall
	creatorShare := pricing.CreatorEarning(amount)
	platformFee := pricing.PlatformFee(amount)

	// Deduct from the caller's wallet.
	_, err = s.walletSvc.Deduct(ctx, callerUserID, amount, botID,
		fmt.Sprintf("Bot call: %s", bot.Name))
	if err != nil {
		return nil, fmt.Errorf("deduct caller: %w", err)
	}

	// Credit the creator's wallet with their share.
	if creatorShare.GreaterThan(decimal.Zero) && bot.UserID != callerUserID {
		_, err = s.walletSvc.Credit(ctx, bot.UserID, creatorShare, botID,
			fmt.Sprintf("Earnings from bot: %s", bot.Name))
		if err != nil {
			s.logger.Error("failed to credit creator earnings",
				zap.String("creator_id", bot.UserID),
				zap.String("bot_id", botID),
				zap.String("amount", creatorShare.String()),
				zap.Error(err),
			)
		}
	}

	s.logger.Debug("bot call billed",
		zap.String("caller", callerUserID),
		zap.String("bot_id", botID),
		zap.String("amount", amount.String()),
		zap.String("creator_share", creatorShare.String()),
		zap.String("platform_fee", platformFee.String()),
	)

	return &BillResult{
		Charged:      true,
		Amount:       amount,
		CreatorShare: creatorShare,
		PlatformFee:  platformFee,
	}, nil
}

// SetPricing creates or updates the pricing configuration for a bot.
func (s *BillingService) SetPricing(ctx context.Context, botID string, pricing *model.BotPricing) error {
	existing, err := s.pricingRepo.FindByBotID(ctx, botID)
	if err == nil {
		existing.Mode = pricing.Mode
		existing.PricePerCall = pricing.PricePerCall
		existing.MonthlyPrice = pricing.MonthlyPrice
		existing.PlatformRate = pricing.PlatformRate
		existing.CreatorRate = pricing.CreatorRate
		existing.FreeCallsPerDay = pricing.FreeCallsPerDay
		return s.pricingRepo.Update(ctx, existing)
	}

	pricing.BotID = botID
	return s.pricingRepo.Create(ctx, pricing)
}

// GetPricing returns the pricing configuration for a bot.
func (s *BillingService) GetPricing(ctx context.Context, botID string) (*model.BotPricing, error) {
	return s.pricingRepo.FindByBotID(ctx, botID)
}
