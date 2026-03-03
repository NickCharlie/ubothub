package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AstrBotAdapter integrates with AstrBot via its HTTP API.
// AstrBot HTTP API: POST /api/v1/message (chat), POST /api/v1/im/message (IM push)
// Auth: X-API-Key header or Authorization: Bearer <api_key>
type AstrBotAdapter struct {
	client *http.Client
}

// NewAstrBotAdapter creates a new AstrBot adapter using the provided shared HTTP client.
func NewAstrBotAdapter(client *http.Client) *AstrBotAdapter {
	return &AstrBotAdapter{client: client}
}

func (a *AstrBotAdapter) Framework() string {
	return "astrbot"
}

// astrBotWebhookPayload represents AstrBot's forwarded message format.
type astrBotWebhookPayload struct {
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Sender    astrBotSender          `json:"sender"`
	Group     *astrBotGroup          `json:"group,omitempty"`
	Platform  string                 `json:"platform"`
	MessageID string                 `json:"message_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type astrBotSender struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

type astrBotGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a *AstrBotAdapter) ParseMessage(rawPayload []byte) (*BotMessage, error) {
	var payload astrBotWebhookPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, fmt.Errorf("parse astrbot payload: %w", err)
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
			Nickname: payload.Sender.Nickname,
		},
	}

	if payload.Sender.Nickname == "" {
		msg.Sender.Nickname = payload.Sender.Name
	}

	if payload.Group != nil {
		msg.Group = &Group{
			GroupID:   payload.Group.ID,
			GroupName: payload.Group.Name,
		}
	}

	if msg.Metadata == nil {
		msg.Metadata = make(map[string]interface{})
	}
	if payload.SessionID != "" {
		msg.Metadata["session_id"] = payload.SessionID
	}

	return msg, nil
}

func (a *AstrBotAdapter) ValidateWebhook(r *http.Request, accessToken string) error {
	// AstrBot webhook validation: check the bot access token in header.
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

// astrBotSendRequest represents the AstrBot /api/v1/im/message request body.
type astrBotSendRequest struct {
	UMO     string      `json:"umo"`
	Message interface{} `json:"message"`
}

// SendMessage sends a message back to AstrBot via its HTTP API.
// webhookURL should be the AstrBot API base URL (e.g., http://localhost:6185).
// The bot config should contain the UMO (unified message origin) and API key.
func (a *AstrBotAdapter) SendMessage(ctx context.Context, webhookURL string, msg *OutboundMessage) error {
	umo, _ := msg.Metadata["umo"].(string)
	apiKey, _ := msg.Metadata["api_key"].(string)

	body := astrBotSendRequest{
		UMO:     umo,
		Message: msg.Content,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal send request: %w", err)
	}

	endpoint := webhookURL + "/api/v1/im/message"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("send message to astrbot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("astrbot returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// AstrBotChatRequest represents the AstrBot /api/v1/message request body.
type AstrBotChatRequest struct {
	Message   interface{} `json:"message"`
	Platform  string      `json:"platform"`
	UserID    string      `json:"user_id"`
	Nickname  string      `json:"nickname"`
	SessionID string      `json:"session_id,omitempty"`
	Timeout   int         `json:"timeout,omitempty"`
}

// AstrBotChatResponse represents the AstrBot /api/v1/message response.
type AstrBotChatResponse struct {
	Success   bool                     `json:"success"`
	Response  []map[string]interface{} `json:"response"`
	EventID   string                   `json:"event_id"`
	SessionID string                   `json:"session_id"`
	Timestamp float64                  `json:"timestamp"`
	Error     string                   `json:"error,omitempty"`
}

// Chat sends a user message to AstrBot's HTTP chat API and returns the response.
// This is the primary method for user→AstrBot→response flow.
// baseURL: AstrBot API base URL (e.g., http://localhost:6185)
// apiKey: AstrBot auth token
// platform: platform identifier (e.g., "ubothub")
func (a *AstrBotAdapter) Chat(ctx context.Context, baseURL, apiKey, platform, userID, nickname, message, sessionID string) (*AstrBotChatResponse, error) {
	body := AstrBotChatRequest{
		Message:   message,
		Platform:  platform,
		UserID:    userID,
		Nickname:  nickname,
		SessionID: sessionID,
		Timeout:   30,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	endpoint := baseURL + "/api/v1/message"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chat with astrbot: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("astrbot returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp AstrBotChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &chatResp, nil
}
