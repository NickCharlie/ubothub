package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// EmailSendPayload carries data required to send an email.
type EmailSendPayload struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	UserID   string `json:"user_id"`
	Template string `json:"template"`
}

// NewEmailSendTask creates a new email send task.
func NewEmailSendTask(payload EmailSendPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal email send payload: %w", err)
	}
	return asynq.NewTask(TypeEmailSend, data, asynq.Queue(QueueCritical), asynq.MaxRetry(3)), nil
}

// AssetProcessPayload carries data required for asset post-processing.
type AssetProcessPayload struct {
	AssetID  string `json:"asset_id"`
	UserID   string `json:"user_id"`
	FileKey  string `json:"file_key"`
	Category string `json:"category"`
	Format   string `json:"format"`
}

// NewAssetProcessTask creates a new asset processing task.
func NewAssetProcessTask(payload AssetProcessPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal asset process payload: %w", err)
	}
	return asynq.NewTask(TypeAssetProcess, data, asynq.Queue(QueueDefault), asynq.MaxRetry(2)), nil
}

// BotHealthCheckPayload carries data for a bot health check ping.
type BotHealthCheckPayload struct {
	BotID      string `json:"bot_id"`
	WebhookURL string `json:"webhook_url"`
	Framework  string `json:"framework"`
}

// NewBotHealthCheckTask creates a new bot health check task.
func NewBotHealthCheckTask(payload BotHealthCheckPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal bot health check payload: %w", err)
	}
	return asynq.NewTask(TypeBotHealthCheck, data, asynq.Queue(QueueLow), asynq.MaxRetry(1)), nil
}

// MessageDispatchPayload carries data for dispatching a message to an AstrBot instance.
type MessageDispatchPayload struct {
	BotID      string                 `json:"bot_id"`
	WebhookURL string                 `json:"webhook_url"`
	Framework  string                 `json:"framework"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewMessageDispatchTask creates a new message dispatch task.
func NewMessageDispatchTask(payload MessageDispatchPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal message dispatch payload: %w", err)
	}
	return asynq.NewTask(TypeMessageDispatch, data, asynq.Queue(QueueCritical), asynq.MaxRetry(3)), nil
}

// MessageLogPayload carries data for persisting a message log entry.
type MessageLogPayload struct {
	BotID           string                 `json:"bot_id"`
	Direction       string                 `json:"direction"`
	Content         string                 `json:"content"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ActionTriggered string                 `json:"action_triggered,omitempty"`
}

// NewMessageLogTask creates a new message log persistence task.
func NewMessageLogTask(payload MessageLogPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal message log payload: %w", err)
	}
	return asynq.NewTask(TypeMessageLog, data, asynq.Queue(QueueLow), asynq.MaxRetry(2)), nil
}
