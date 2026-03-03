package circuitbreaker

import (
	"net/url"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"
)

// Registry manages per-host circuit breakers for outbound HTTP requests.
// Each unique host (from the webhook URL) gets its own circuit breaker to
// isolate failures between different AstrBot instances.
type Registry struct {
	breakers map[string]*gobreaker.CircuitBreaker[[]byte]
	mu       sync.RWMutex
	settings gobreaker.Settings
}

// NewRegistry creates a new circuit breaker registry with sensible defaults
// for external API calls to AstrBot instances.
func NewRegistry() *Registry {
	return &Registry{
		breakers: make(map[string]*gobreaker.CircuitBreaker[[]byte]),
		settings: gobreaker.Settings{
			MaxRequests: 2,                // allow 2 requests in half-open state
			Interval:    60 * time.Second, // reset failure count after 60s of no failures
			Timeout:     30 * time.Second, // stay open for 30s before transitioning to half-open
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 5
			},
		},
	}
}

// Get returns the circuit breaker for the given URL's host.
// Creates a new breaker if one doesn't exist yet.
func (r *Registry) Get(rawURL string) *gobreaker.CircuitBreaker[[]byte] {
	host := extractHost(rawURL)

	r.mu.RLock()
	cb, ok := r.breakers[host]
	r.mu.RUnlock()
	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock.
	if cb, ok = r.breakers[host]; ok {
		return cb
	}

	settings := r.settings
	settings.Name = host
	cb = gobreaker.NewCircuitBreaker[[]byte](settings)
	r.breakers[host] = cb
	return cb
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}
