package handler

import (
	"errors"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/adapter"
	"github.com/NickCharlie/ubothub/backend/internal/event"
	"github.com/NickCharlie/ubothub/backend/internal/moderation"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	"github.com/NickCharlie/ubothub/backend/pkg/response"
	"go.uber.org/zap"
)

// GatewayHandler handles bot webhook and message ingestion endpoints.
type GatewayHandler struct {
	botSvc         *service.BotService
	adapterFactory *adapter.Factory
	eventBus       *event.Bus
	moderator      moderation.Service
	logger         *zap.Logger
}

// NewGatewayHandler creates a new gateway handler.
func NewGatewayHandler(
	botSvc *service.BotService,
	adapterFactory *adapter.Factory,
	eventBus *event.Bus,
	moderator moderation.Service,
	logger *zap.Logger,
) *GatewayHandler {
	return &GatewayHandler{
		botSvc:         botSvc,
		adapterFactory: adapterFactory,
		eventBus:       eventBus,
		moderator:      moderator,
		logger:         logger,
	}
}

// Webhook handles POST /api/v1/gateway/webhook/:token.
// Bot frameworks send messages to this endpoint using the bot's access token.
func (h *GatewayHandler) Webhook(c *gin.Context) {
	accessToken := c.Param("token")
	if accessToken == "" {
		response.Error(c, errcode.ErrBotTokenInvalid)
		return
	}

	bot, err := h.botSvc.GetBotByAccessToken(c.Request.Context(), accessToken)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			response.Error(c, errcode.ErrBotTokenInvalid)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	adpt, err := h.adapterFactory.Get(bot.Framework)
	if err != nil {
		h.logger.Error("adapter not found", zap.String("framework", bot.Framework), zap.Error(err))
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20)) // 1MB limit
	if err != nil {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	msg, err := adpt.ParseMessage(body)
	if err != nil {
		h.logger.Warn("failed to parse webhook message",
			zap.String("bot_id", bot.ID),
			zap.Error(err),
		)
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	// Content moderation check.
	if msg.Content != "" {
		result, err := h.moderator.CheckText(c.Request.Context(), msg.Content)
		if err == nil && !result.Pass {
			h.logger.Warn("webhook message blocked by content moderation",
				zap.String("bot_id", bot.ID),
				zap.Strings("labels", result.Labels),
			)
			response.Error(c, errcode.ErrContentViolation)
			return
		}
	}

	h.eventBus.Publish(c.Request.Context(), event.Event{
		Type: event.BotMessageReceived,
		Payload: event.BotMessageEvent{
			BotID:    bot.ID,
			Content:  msg.Content,
			Platform: msg.Platform,
			Sender: event.MessageSender{
				UserID:   msg.Sender.UserID,
				Nickname: msg.Sender.Nickname,
			},
			Metadata:  msg.Metadata,
			Timestamp: msg.Timestamp,
		},
		Timestamp: time.Now().Unix(),
	})

	h.logger.Debug("webhook message received",
		zap.String("bot_id", bot.ID),
		zap.String("framework", bot.Framework),
	)

	response.OKWithMessage(c, "message received")
}

// Message handles POST /api/v1/gateway/message.
// Authenticated via Bearer token in Authorization header.
func (h *GatewayHandler) Message(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	var accessToken string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}
	if accessToken == "" {
		response.Error(c, errcode.ErrBotTokenInvalid)
		return
	}

	bot, err := h.botSvc.GetBotByAccessToken(c.Request.Context(), accessToken)
	if err != nil {
		if errors.Is(err, service.ErrBotNotFound) {
			response.Error(c, errcode.ErrBotTokenInvalid)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	adpt, err := h.adapterFactory.Get(bot.Framework)
	if err != nil {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	msg, err := adpt.ParseMessage(body)
	if err != nil {
		response.Error(c, errcode.ErrBadRequest)
		return
	}

	// Content moderation check.
	if msg.Content != "" {
		result, err := h.moderator.CheckText(c.Request.Context(), msg.Content)
		if err == nil && !result.Pass {
			h.logger.Warn("message blocked by content moderation",
				zap.String("bot_id", bot.ID),
				zap.Strings("labels", result.Labels),
			)
			response.Error(c, errcode.ErrContentViolation)
			return
		}
	}

	h.eventBus.Publish(c.Request.Context(), event.Event{
		Type: event.BotMessageReceived,
		Payload: event.BotMessageEvent{
			BotID:    bot.ID,
			Content:  msg.Content,
			Platform: msg.Platform,
			Sender: event.MessageSender{
				UserID:   msg.Sender.UserID,
				Nickname: msg.Sender.Nickname,
			},
			Metadata:  msg.Metadata,
			Timestamp: msg.Timestamp,
		},
		Timestamp: time.Now().Unix(),
	})

	response.OKWithMessage(c, "message received")
}
