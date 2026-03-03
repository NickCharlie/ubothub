package queue

import (
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Client wraps the asynq client for task enqueueing.
type Client struct {
	client *asynq.Client
	logger *zap.Logger
}

// NewClient creates a new asynq task client.
func NewClient(redisAddr, redisPassword string, logger *zap.Logger) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})
	return &Client{client: client, logger: logger}
}

// Enqueue submits a task to the queue for async processing.
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) error {
	info, err := c.client.Enqueue(task, opts...)
	if err != nil {
		c.logger.Warn("failed to enqueue task",
			zap.String("type", task.Type()),
			zap.Error(err),
		)
		return fmt.Errorf("enqueue task %s: %w", task.Type(), err)
	}

	c.logger.Debug("task enqueued",
		zap.String("id", info.ID),
		zap.String("type", task.Type()),
		zap.String("queue", info.Queue),
	)
	return nil
}

// Close releases the underlying asynq client resources.
func (c *Client) Close() error {
	return c.client.Close()
}
