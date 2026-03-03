package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"go.uber.org/zap"
)

// MessageLogHandler processes message log persistence tasks.
type MessageLogHandler struct {
	repo   repository.MessageLogRepository
	logger *zap.Logger
}

// NewMessageLogHandler creates a new message log task handler.
func NewMessageLogHandler(repo repository.MessageLogRepository, logger *zap.Logger) *MessageLogHandler {
	return &MessageLogHandler{repo: repo, logger: logger}
}

// ProcessTask handles a message log persistence task.
func (h *MessageLogHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p MessageLogPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal message log payload: %w", err)
	}

	metadata, _ := json.Marshal(p.Metadata)
	if len(metadata) == 0 {
		metadata = []byte("{}")
	}

	log := &model.MessageLog{
		ID:              xid.New().String(),
		BotID:           p.BotID,
		Direction:       p.Direction,
		Content:         p.Content,
		Metadata:        string(metadata),
		ActionTriggered: p.ActionTriggered,
	}

	if err := h.repo.Create(ctx, log); err != nil {
		h.logger.Warn("message log persistence failed",
			zap.String("bot_id", p.BotID),
			zap.Error(err),
		)
		return fmt.Errorf("persist message log: %w", err)
	}

	h.logger.Debug("message log persisted",
		zap.String("bot_id", p.BotID),
		zap.String("direction", p.Direction),
	)
	return nil
}

// MessageLogPayload is the task payload for message log persistence.
type MessageLogPayload struct {
	BotID           string                 `json:"bot_id"`
	Direction       string                 `json:"direction"`
	Content         string                 `json:"content"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ActionTriggered string                 `json:"action_triggered,omitempty"`
}
