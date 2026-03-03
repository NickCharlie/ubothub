package adapter

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/NickCharlie/ubothub/backend/pkg/circuitbreaker"
	"github.com/NickCharlie/ubothub/backend/pkg/retry"
)

// Factory manages adapter registration and lookup.
type Factory struct {
	adapters map[string]BotAdapter
	mu       sync.RWMutex
}

// NewFactory creates a new adapter factory with default adapters registered.
// The shared HTTP client is used by all adapters for connection pooling.
// Each adapter is wrapped with circuit breaker and retry for resilient outbound calls.
func NewFactory(client *http.Client) *Factory {
	cbRegistry := circuitbreaker.NewRegistry()
	retryCfg := retry.DefaultConfig()

	f := &Factory{
		adapters: make(map[string]BotAdapter),
	}
	f.Register(NewResilientAdapter(NewAstrBotAdapter(client), cbRegistry, retryCfg))
	f.Register(NewResilientAdapter(NewWebhookAdapter(client), cbRegistry, retryCfg))
	return f
}

// Register adds a bot adapter to the factory.
func (f *Factory) Register(adapter BotAdapter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adapters[adapter.Framework()] = adapter
}

// Get returns the adapter for the given framework.
func (f *Factory) Get(framework string) (BotAdapter, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	adapter, ok := f.adapters[framework]
	if !ok {
		// Fall back to generic webhook adapter for unknown frameworks.
		adapter, ok = f.adapters["custom"]
		if !ok {
			return nil, fmt.Errorf("no adapter found for framework: %s", framework)
		}
	}
	return adapter, nil
}
