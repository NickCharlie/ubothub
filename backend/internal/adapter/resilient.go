package adapter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sony/gobreaker/v2"
	"github.com/NickCharlie/ubothub/backend/pkg/circuitbreaker"
	"github.com/NickCharlie/ubothub/backend/pkg/retry"
)

// ResilientAdapter wraps a BotAdapter with circuit breaker and retry logic
// for outbound SendMessage calls. Inbound parsing and webhook validation
// are delegated directly without resilience wrappers.
type ResilientAdapter struct {
	inner    BotAdapter
	registry *circuitbreaker.Registry
	retryCfg retry.Config
}

// NewResilientAdapter wraps the given adapter with per-host circuit breaker and retry.
func NewResilientAdapter(inner BotAdapter, registry *circuitbreaker.Registry, retryCfg retry.Config) *ResilientAdapter {
	return &ResilientAdapter{
		inner:    inner,
		registry: registry,
		retryCfg: retryCfg,
	}
}

// Inner returns the underlying unwrapped adapter.
func (a *ResilientAdapter) Inner() BotAdapter {
	return a.inner
}

func (a *ResilientAdapter) Framework() string {
	return a.inner.Framework()
}

func (a *ResilientAdapter) ParseMessage(rawPayload []byte) (*BotMessage, error) {
	return a.inner.ParseMessage(rawPayload)
}

func (a *ResilientAdapter) ValidateWebhook(r *http.Request, accessToken string) error {
	return a.inner.ValidateWebhook(r, accessToken)
}

// SendMessage sends a message with circuit breaker protection and retry logic.
// The circuit breaker is per-host, so a failing AstrBot instance won't affect others.
func (a *ResilientAdapter) SendMessage(ctx context.Context, webhookURL string, msg *OutboundMessage) error {
	cb := a.registry.Get(webhookURL)

	return retry.Do(ctx, a.retryCfg, func(ctx context.Context) error {
		_, err := cb.Execute(func() ([]byte, error) {
			if err := a.inner.SendMessage(ctx, webhookURL, msg); err != nil {
				return nil, err
			}
			return nil, nil
		})
		if err != nil {
			// If the circuit breaker is open, don't retry — fail fast.
			if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
				return fmt.Errorf("circuit breaker open for %s: %w", webhookURL, err)
			}
			return err
		}
		return nil
	})
}
