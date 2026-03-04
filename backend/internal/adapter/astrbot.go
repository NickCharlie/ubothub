package adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// AstrBotChatRequest represents the AstrBot /api/v1/chat request body.
type AstrBotChatRequest struct {
	Message   interface{} `json:"message"`
	Username  string      `json:"username"`
	SessionID string      `json:"session_id,omitempty"`
}

// AstrBotChatResponse holds the accumulated result from AstrBot's SSE chat stream.
type AstrBotChatResponse struct {
	Text      string `json:"text"`
	SessionID string `json:"session_id"`
}

// Chat sends a user message to AstrBot's Open API (/api/v1/chat) and collects
// the SSE stream into a single text response.
// baseURL: AstrBot base URL (e.g., http://localhost:6185)
// apiKey: AstrBot API key (raw, not hashed)
func (a *AstrBotAdapter) Chat(ctx context.Context, baseURL, apiKey, username, sessionID, message string) (*AstrBotChatResponse, error) {
	body := AstrBotChatRequest{
		Message:   message,
		Username:  username,
		SessionID: sessionID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	endpoint := baseURL + "/api/v1/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chat with astrbot: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("astrbot returned %d: %s", resp.StatusCode, string(errBody))
	}

	// Parse SSE stream. AstrBot streams "data: {json}\n\n" events.
	// Collect all "plain" text chunks until "end" event.
	result := &AstrBotChatResponse{}
	return result, a.parseSSEStream(resp.Body, result)
}

// parseSSEStream reads an SSE text/event-stream body and accumulates text content.
func (a *AstrBotAdapter) parseSSEStream(body io.Reader, result *AstrBotChatResponse) error {
	scanner := bufio.NewScanner(body)
	// Allow large SSE lines (up to 256KB).
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := line[6:] // strip "data: " prefix

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		msgType, _ := event["type"].(string)

		// Capture session_id from the initial event.
		if msgType == "session_id" {
			if sid, ok := event["session_id"].(string); ok {
				result.SessionID = sid
			}
			continue
		}

		// End of stream.
		if msgType == "end" || msgType == "complete" {
			break
		}

		// Error from AstrBot.
		if msgType == "error" {
			errMsg, _ := event["data"].(string)
			if errMsg == "" {
				errMsg = "unknown error"
			}
			return fmt.Errorf("astrbot error: %s", errMsg)
		}

		// Accumulate plain text chunks (skip tool_call, reasoning, etc.).
		if msgType == "plain" {
			chainType, _ := event["chain_type"].(string)
			if chainType == "tool_call" || chainType == "tool_call_result" || chainType == "reasoning" || chainType == "agent_stats" {
				continue
			}
			text, _ := event["data"].(string)
			result.Text += text
		}
	}

	return scanner.Err()
}
