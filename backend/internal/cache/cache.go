package cache

import (
	"context"
	"time"
)

// Cache defines the interface for cache operations. Implementations
// must be safe for concurrent use.
type Cache interface {
	// Get retrieves a value by key and unmarshals it into dest.
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value with the given key and TTL. A zero TTL means
	// the key does not expire.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes one or more keys from the cache.
	Delete(ctx context.Context, keys ...string) error

	// Exists checks whether a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}
