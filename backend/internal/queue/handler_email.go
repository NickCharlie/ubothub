package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/NickCharlie/ubothub/backend/internal/queue/tasks"
	"github.com/NickCharlie/ubothub/backend/pkg/email"
	"go.uber.org/zap"
)

// EmailHandler processes email sending tasks.
type EmailHandler struct {
	sender *email.Sender
	logger *zap.Logger
}

// NewEmailHandler creates a new email task handler.
func NewEmailHandler(sender *email.Sender, logger *zap.Logger) *EmailHandler {
	return &EmailHandler{sender: sender, logger: logger}
}

// ProcessTask handles an email send task.
func (h *EmailHandler) ProcessTask(_ context.Context, t *asynq.Task) error {
	var p tasks.EmailSendPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal email payload: %w", err)
	}

	if err := h.sender.Send(p.To, p.Subject, p.HTMLBody); err != nil {
		h.logger.Warn("email send failed, will retry",
			zap.String("to", p.To),
			zap.String("template", p.Template),
			zap.Error(err),
		)
		return fmt.Errorf("send email to %s: %w", p.To, err)
	}

	h.logger.Info("email sent successfully",
		zap.String("to", p.To),
		zap.String("template", p.Template),
		zap.String("user_id", p.UserID),
	)
	return nil
}
