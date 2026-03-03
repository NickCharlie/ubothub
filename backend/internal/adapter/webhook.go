package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookAdapter is a generic adapter for custom bot frameworks using HTTP webhooks.
type WebhookAdapter struct {
	client *http.Client
}

// NewWebhookAdapter creates a new generic webhook adapter.
func NewWebhookAdapter() *WebhookAdapter {
	return &WebhookAdapter{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *WebhookAdapter) Framework() string {
	return "custom"
}

// genericWebhookPayload represents the standardized webhook payload format.
type genericWebhookPayload struct {
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Sender    genericSender          `json:"sender"`
	Group     *genericGroup          `json:"group,omitempty"`
	Platform  string                 `json:"platform"`
	MessageID string                 `json:"message_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type genericSender struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type genericGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a *WebhookAdapter) ParseMessage(rawPayload []byte) (*BotMessage, error) {
	var payload genericWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("parse webhook payload: %w", err)
	}

	msg := &BotMessage{
		Type:      payload.Type,
		Content:   payload.Content,
		Platform:  payload.Platform,
		MessageID: payload.MessageID,
		Timestamp: payload.Timestamp,
		Metadata:  payload.Metadata,
		Sender: Sender{
			UserID:   payload.Sender.ID,
			Nickname: payload.Sender.Name,
		},
	}

	if payload.Group != nil {
		msg.Group = &Group{
			GroupID:   payload.Group.ID,
			GroupName: payload.Group.Name,
		}
	}

	return msg, nil
}

func (a *WebhookAdapter) ValidateWebhook(r *http.Request, accessToken string) error {
	token := r.Header.Get("X-Bot-Token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}
	if token != accessToken {
		return fmt.Errorf("invalid bot access token")
	}
	return nil
}

func (a *WebhookAdapter) SendMessage(ctx context.Context, webhookURL string, msg *OutboundMessage) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	jsonBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal outbound message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
