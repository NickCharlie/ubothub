package adapter

import (
	"context"
	"net/http"
)

// BotMessage represents a normalized incoming message from any bot framework.
type BotMessage struct {
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Sender    Sender                 `json:"sender"`
	Group     *Group                 `json:"group,omitempty"`
	Platform  string                 `json:"platform"`
	MessageID string                 `json:"message_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Sender represents the message sender.
type Sender struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

// Group represents group info for group messages.
type Group struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// OutboundMessage represents a message to send back through the bot framework.
type OutboundMessage struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BotAdapter defines the interface for integrating with different bot frameworks.
type BotAdapter interface {
	// Framework returns the adapter's framework identifier.
	Framework() string

	// ParseMessage normalizes a raw webhook/API payload into a BotMessage.
	ParseMessage(rawPayload []byte) (*BotMessage, error)

	// ValidateWebhook verifies the authenticity of an incoming webhook request.
	ValidateWebhook(r *http.Request, accessToken string) error

	// SendMessage sends a message back through the bot framework.
	SendMessage(ctx context.Context, webhookURL string, msg *OutboundMessage) error
}
