package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/payment"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	"github.com/NickCharlie/ubothub/backend/pkg/response"
)

// WalletHandler handles wallet and billing HTTP endpoints.
type WalletHandler struct {
	walletSvc  *service.WalletService
	billingSvc *service.BillingService
	paymentPvd payment.Provider
}

// NewWalletHandler creates a new wallet handler.
func NewWalletHandler(walletSvc *service.WalletService, billingSvc *service.BillingService, paymentPvd payment.Provider) *WalletHandler {
	return &WalletHandler{walletSvc: walletSvc, billingSvc: billingSvc, paymentPvd: paymentPvd}
}

// GetWallet handles GET /api/v1/wallet.
// @Summary Get wallet balance
// @Description Returns the current user's wallet balance and details.
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Router /wallet [get]
func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID := c.GetString("user_id")
	wallet, err := h.walletSvc.GetOrCreateWallet(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OK(c, gin.H{
		"wallet_id":         wallet.ID,
		"balance":           wallet.Balance.String(),
		"frozen":            wallet.Frozen.String(),
		"available_balance": wallet.AvailableBalance().String(),
		"currency":          wallet.Currency,
	})
}

// TopUp handles POST /api/v1/wallet/topup.
// @Summary Initiate wallet top-up
// @Description Create a payment order for wallet top-up via specified channel.
// @Tags Wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object true "Top-up payload (amount, channel)"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /wallet/topup [post]
func (h *WalletHandler) TopUp(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Amount  string `json:"amount" binding:"required"`
		Channel string `json:"channel" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.Error(c, errcode.ErrInvalidAmount)
		return
	}

	orderID := xid.New().String()
	result, err := h.paymentPvd.CreateOrder(c.Request.Context(), payment.CreateOrderRequest{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      amount,
		Description: "UBotHub wallet top-up",
		Channel:     req.Channel,
		NotifyURL:   "/api/v1/payment/notify",
	})
	if err != nil {
		response.Error(c, errcode.ErrPaymentFailed)
		return
	}

	response.OK(c, gin.H{
		"order_id":   result.OrderID,
		"pay_url":    result.PayURL,
		"qr_code":    result.QRCodeURL,
		"channel":    result.Channel,
	})
}

// Transactions handles GET /api/v1/wallet/transactions.
// @Summary Get transaction history
// @Description Returns paginated transaction history for the current user.
// @Tags Wallet
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param type query string false "Filter by transaction type"
// @Success 200 {object} response.CommonResponse
// @Router /wallet/transactions [get]
func (h *WalletHandler) Transactions(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	txType := c.Query("type")

	txns, total, err := h.walletSvc.GetTransactions(c.Request.Context(), userID, page, pageSize, txType)
	if err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.OKPaged(c, txns, total, page, pageSize)
}

// SetBotPricing handles PUT /api/v1/bots/:id/pricing.
// @Summary Set bot pricing
// @Description Configure the pricing model for a bot (creator only).
// @Tags Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Param body body object true "Pricing payload (mode, price_per_call, monthly_price, free_calls_per_day)"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /bots/{id}/pricing [put]
func (h *WalletHandler) SetBotPricing(c *gin.Context) {
	botID := c.Param("id")

	var req struct {
		Mode            string `json:"mode" binding:"required"`
		PricePerCall    string `json:"price_per_call"`
		MonthlyPrice    string `json:"monthly_price"`
		FreeCallsPerDay int    `json:"free_calls_per_day"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	pricePerCall, _ := decimal.NewFromString(req.PricePerCall)
	monthlyPrice, _ := decimal.NewFromString(req.MonthlyPrice)

	pricing := &model.BotPricing{
		ID:              xid.New().String(),
		Mode:            req.Mode,
		PricePerCall:    pricePerCall,
		MonthlyPrice:    monthlyPrice,
		PlatformRate:    decimal.NewFromFloat(0.20),
		CreatorRate:     decimal.NewFromFloat(0.80),
		FreeCallsPerDay: req.FreeCallsPerDay,
	}

	if err := h.billingSvc.SetPricing(c.Request.Context(), botID, pricing); err != nil {
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	response.OKWithMessage(c, "pricing updated")
}

// GetBotPricing handles GET /api/v1/bots/:id/pricing.
// @Summary Get bot pricing
// @Description Returns the pricing configuration for a bot.
// @Tags Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bot ID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /bots/{id}/pricing [get]
func (h *WalletHandler) GetBotPricing(c *gin.Context) {
	botID := c.Param("id")

	pricing, err := h.billingSvc.GetPricing(c.Request.Context(), botID)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			response.Error(c, errcode.ErrBotNotFound)
			return
		}
		// No pricing configured means free.
		response.OK(c, gin.H{
			"mode":               model.PricingModeFree,
			"price_per_call":     "0",
			"monthly_price":      "0",
			"platform_rate":      "0.20",
			"creator_rate":       "0.80",
			"free_calls_per_day": 0,
		})
		return
	}

	response.OK(c, pricing)
}
