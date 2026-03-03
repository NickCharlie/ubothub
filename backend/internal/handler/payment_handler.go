package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/payment"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"go.uber.org/zap"
)

// PaymentHandler handles payment notification callbacks from WeChat/Alipay.
// These endpoints receive async notifications when payment status changes.
type PaymentHandler struct {
	registry  *payment.Registry
	walletSvc *service.WalletService
	logger    *zap.Logger
}

// NewPaymentHandler creates a new payment notification handler.
func NewPaymentHandler(registry *payment.Registry, walletSvc *service.WalletService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		registry:  registry,
		walletSvc: walletSvc,
		logger:    logger,
	}
}

// WechatNotify handles POST /api/v1/payment/notify/wechat.
// @Summary WeChat payment notification
// @Description Receives WeChat Pay V3 async payment notifications with double verification.
// @Tags Payment
// @Accept json
// @Produce json
// @Success 200 {object} object
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Router /payment/notify/wechat [post]
func (h *PaymentHandler) WechatNotify(c *gin.Context) {
	h.handleNotify(c, "wechat")
}

// AlipayNotify handles POST /api/v1/payment/notify/alipay.
// @Summary Alipay payment notification
// @Description Receives Alipay async payment notifications with double verification.
// @Tags Payment
// @Accept json
// @Produce plain
// @Success 200 {string} string "success"
// @Failure 400 {string} string
// @Failure 403 {string} string
// @Router /payment/notify/alipay [post]
func (h *PaymentHandler) AlipayNotify(c *gin.Context) {
	h.handleNotify(c, "alipay")
}

// handleNotify processes a payment notification with double verification.
func (h *PaymentHandler) handleNotify(c *gin.Context, providerName string) {
	provider, err := h.registry.GetByName(providerName)
	if err != nil {
		h.logger.Error("payment provider not found", zap.String("provider", providerName), zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Step 1: Verify notification signature using the raw HTTP request.
	valid, err := provider.VerifyNotify(c.Request.Context(), c.Request)
	if err != nil || !valid {
		h.logger.Warn("payment notification signature verification failed",
			zap.String("provider", providerName),
			zap.Error(err),
		)
		c.Status(http.StatusForbidden)
		return
	}

	// Step 2: Parse notification payload from the HTTP request.
	result, err := provider.ParseNotify(c.Request.Context(), c.Request)
	if err != nil {
		h.logger.Error("failed to parse payment notification",
			zap.String("provider", providerName),
			zap.Error(err),
		)
		c.Status(http.StatusBadRequest)
		return
	}

	// Step 3: Double verification - query order status from official API.
	queryResult, err := provider.QueryOrder(c.Request.Context(), result.OrderID)
	if err != nil {
		h.logger.Warn("payment order query for double verification failed",
			zap.String("provider", providerName),
			zap.String("order_id", result.OrderID),
			zap.Error(err),
		)
		// Continue with notification result if query fails (degraded mode).
	} else if queryResult.Status != payment.OrderStatusPaid {
		h.logger.Warn("payment notification status mismatch with query result",
			zap.String("provider", providerName),
			zap.String("order_id", result.OrderID),
			zap.String("notify_status", result.Status),
			zap.String("query_status", queryResult.Status),
		)
		// Use query result as source of truth.
		result.Status = queryResult.Status
		result.Amount = queryResult.Amount
	}

	if result.Status != payment.OrderStatusPaid {
		h.logger.Debug("payment notification received but not paid",
			zap.String("order_id", result.OrderID),
			zap.String("status", result.Status),
		)
		h.respondSuccess(c, providerName)
		return
	}

	// Step 4: Log the confirmed payment.
	// In production, this should look up a pending_orders table to map
	// the order_id to a user_id and call walletSvc.TopUp accordingly.
	h.logger.Info("payment confirmed",
		zap.String("provider", providerName),
		zap.String("order_id", result.OrderID),
		zap.String("external_id", result.ExternalID),
		zap.String("amount", result.Amount.String()),
	)

	h.respondSuccess(c, providerName)
}

// respondSuccess sends the appropriate success response based on provider.
func (h *PaymentHandler) respondSuccess(c *gin.Context, providerName string) {
	switch providerName {
	case "wechat":
		// WeChat V3 expects JSON response.
		c.JSON(http.StatusOK, gin.H{
			"code":    "SUCCESS",
			"message": "OK",
		})
	case "alipay":
		// Alipay expects plain text "success".
		c.String(http.StatusOK, "success")
	default:
		c.Status(http.StatusOK)
	}
}
