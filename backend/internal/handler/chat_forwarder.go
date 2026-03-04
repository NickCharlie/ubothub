package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/NickCharlie/ubothub/backend/internal/adapter"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/internal/ws"
	"go.uber.org/zap"
)

// ChatForwarder bridges WebSocket client messages to bot adapters.
// When a user sends a chat message, it forwards to the appropriate bot
// framework (e.g., AstrBot HTTP API) and returns the response via WebSocket.
type ChatForwarder struct {
	botSvc         *service.BotService
	adapterFactory *adapter.Factory
	logger         *zap.Logger
}

// NewChatForwarder creates a new chat forwarder.
func NewChatForwarder(botSvc *service.BotService, adapterFactory *adapter.Factory, logger *zap.Logger) *ChatForwarder {
	return &ChatForwarder{
		botSvc:         botSvc,
		adapterFactory: adapterFactory,
		logger:         logger,
	}
}

// HandleMessage is the callback registered with Hub.SetMessageHandler.
// It runs in the hub's goroutine, so heavy work is dispatched asynchronously.
func (f *ChatForwarder) HandleMessage(client *ws.Client, msg *ws.InboundMessage) {
	if msg.Type != ws.TypeChat && msg.Type != "message" {
		return
	}

	if msg.Content == "" {
		return
	}

	botID := msg.BotID
	if botID == "" {
		botID = client.RoomID()
	}

	// Dispatch to goroutine to avoid blocking the hub event loop.
	go f.forwardToBot(client, botID, msg.Content, client.UserID())
}

func (f *ChatForwarder) forwardToBot(client *ws.Client, botID, content, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bot, err := f.botSvc.GetBotByID(ctx, botID)
	if err != nil {
		f.sendError(client, botID, "Bot not found or unavailable")
		f.logger.Warn("chat forward: bot not found",
			zap.String("bot_id", botID),
			zap.Error(err),
		)
		return
	}

	// Decrypt bot config to get API key and other settings.
	configMap, err := f.botSvc.DecryptBotConfig(bot.Config)
	if err != nil {
		f.sendError(client, botID, "Failed to read bot configuration")
		return
	}

	switch bot.Framework {
	case "astrbot":
		f.forwardToAstrBot(ctx, client, bot.ID, bot.WebhookURL, configMap, content, userID)
	default:
		f.sendError(client, botID, fmt.Sprintf("Chat not supported for framework: %s", bot.Framework))
	}
}

func (f *ChatForwarder) forwardToAstrBot(
	ctx context.Context,
	client *ws.Client,
	botID, webhookURL string,
	configMap map[string]interface{},
	content, userID string,
) {
	if webhookURL == "" {
		f.sendError(client, botID, "AstrBot API base URL not configured")
		return
	}

	apiKey, _ := configMap["api_key"].(string)
	platform, _ := configMap["platform"].(string)
	if platform == "" {
		platform = "ubothub"
	}

	f.logger.Debug("astrbot chat config",
		zap.String("bot_id", botID),
		zap.String("webhook_url", webhookURL),
		zap.Int("api_key_len", len(apiKey)),
		zap.String("platform", platform),
	)

	// Get the raw AstrBot adapter (unwrap resilient wrapper).
	adpt, err := f.adapterFactory.Get("astrbot")
	if err != nil {
		f.sendError(client, botID, "AstrBot adapter not available")
		return
	}

	// Type-assert to get the Chat method. The resilient adapter wraps the real one.
	var astrBotAdpt *adapter.AstrBotAdapter
	switch a := adpt.(type) {
	case *adapter.AstrBotAdapter:
		astrBotAdpt = a
	case *adapter.ResilientAdapter:
		if inner, ok := a.Inner().(*adapter.AstrBotAdapter); ok {
			astrBotAdpt = inner
		}
	}

	if astrBotAdpt == nil {
		f.sendError(client, botID, "AstrBot adapter unavailable")
		return
	}

	// Build session ID for AstrBot context continuity.
	sessionID := fmt.Sprintf("%s_%s", platform, userID)

	chatResp, err := astrBotAdpt.Chat(ctx, webhookURL, apiKey, platform, userID, "", content, sessionID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			errMsg = "Cannot connect to AstrBot instance. Please check if it's running."
		}
		f.sendError(client, botID, errMsg)
		f.logger.Warn("astrbot chat failed",
			zap.String("bot_id", botID),
			zap.Error(err),
		)
		return
	}

	if !chatResp.Success {
		f.sendError(client, botID, "AstrBot returned an error: "+chatResp.Error)
		return
	}

	// Extract text content from AstrBot response components.
	replyText := extractAstrBotText(chatResp.Response)

	client.Hub().SendToClient(client, &ws.OutboundMessage{
		Type:      ws.TypeBotReply,
		BotID:     botID,
		Content:   replyText,
		Sender:    "AstrBot",
		Timestamp: time.Now().Unix(),
		Data: map[string]interface{}{
			"event_id":   chatResp.EventID,
			"session_id": chatResp.SessionID,
			"raw":        chatResp.Response,
		},
	})
}

func (f *ChatForwarder) sendError(client *ws.Client, botID, errMsg string) {
	client.Hub().SendToClient(client, &ws.OutboundMessage{
		Type:      ws.TypeError,
		BotID:     botID,
		Content:   errMsg,
		Timestamp: time.Now().Unix(),
	})
}

// extractAstrBotText extracts plain text from AstrBot response components.
func extractAstrBotText(components []map[string]interface{}) string {
	var parts []string
	for _, comp := range components {
		content, ok := comp["content"].(map[string]interface{})
		if !ok {
			continue
		}
		data, ok := content["data"].(map[string]interface{})
		if !ok {
			continue
		}
		if text, ok := data["text"].(string); ok && text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		return "(empty response)"
	}
	return strings.Join(parts, "")
}
