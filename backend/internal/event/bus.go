package event

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Event represents an internal application event.
type Event struct {
	Type      string
	Payload   interface{}
	Timestamp int64
}

// Handler processes an event.
type Handler func(ctx context.Context, event Event) error

// Bus is an in-process event bus backed by Go channels and a worker pool.
type Bus struct {
	subscribers map[string][]Handler
	mu          sync.RWMutex
	workerPool  chan struct{}
	logger      *zap.Logger
}

// NewBus creates a new event bus with the given concurrency limit.
func NewBus(concurrency int, logger *zap.Logger) *Bus {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &Bus{
		subscribers: make(map[string][]Handler),
		workerPool:  make(chan struct{}, concurrency),
		logger:      logger,
	}
}

// Subscribe registers a handler for the given event type.
func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
}

// Publish dispatches an event to all subscribers asynchronously.
func (b *Bus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := b.subscribers[event.Type]
	b.mu.RUnlock()

	for _, h := range handlers {
		h := h
		b.workerPool <- struct{}{}
		go func() {
			defer func() { <-b.workerPool }()
			if err := h(ctx, event); err != nil {
				b.logger.Error("event handler failed",
					zap.String("event_type", event.Type),
					zap.Error(err),
				)
			}
		}()
	}
}

// PublishSync dispatches an event to all subscribers synchronously.
func (b *Bus) PublishSync(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers := b.subscribers[event.Type]
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
