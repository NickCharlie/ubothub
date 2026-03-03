package adapter

import (
	"fmt"
	"sync"
)

// Factory manages adapter registration and lookup.
type Factory struct {
	adapters map[string]BotAdapter
	mu       sync.RWMutex
}

// NewFactory creates a new adapter factory with default adapters registered.
func NewFactory() *Factory {
	f := &Factory{
		adapters: make(map[string]BotAdapter),
	}
	f.Register(NewAstrBotAdapter())
	f.Register(NewWebhookAdapter())
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
