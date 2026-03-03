package queue

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/NickCharlie/ubothub/backend/internal/queue/tasks"
	"go.uber.org/zap"
)

// Server wraps the asynq server and mux for task processing.
type Server struct {
	srv    *asynq.Server
	mux    *asynq.ServeMux
	logger *zap.Logger
}

// NewServer creates a new asynq task processing server.
func NewServer(redisAddr, redisPassword string, concurrency int, logger *zap.Logger) *Server {
	if concurrency <= 0 {
		concurrency = 10
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				tasks.QueueCritical: 6,
				tasks.QueueDefault:  3,
				tasks.QueueLow:      1,
			},
			Logger:   newZapAdapter(logger),
			LogLevel: asynq.WarnLevel,
		},
	)

	mux := asynq.NewServeMux()

	return &Server{
		srv:    srv,
		mux:    mux,
		logger: logger,
	}
}

// Register binds a task handler to a task type on the internal mux.
func (s *Server) Register(taskType string, handler asynq.Handler) {
	s.mux.Handle(taskType, handler)
}

// Start begins processing tasks. This call blocks until Stop is called or
// the server encounters an unrecoverable error.
func (s *Server) Start() error {
	s.logger.Info("asynq worker started")
	if err := s.srv.Start(s.mux); err != nil {
		return fmt.Errorf("start asynq server: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the asynq server.
func (s *Server) Stop() {
	s.srv.Stop()
	s.logger.Info("asynq worker stopped")
}
