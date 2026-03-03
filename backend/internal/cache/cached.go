package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// emptyPlaceholder is stored in cache when the data source returns nil,
// preventing cache penetration from repeated queries for non-existent keys.
var emptyPlaceholder = []byte("null")

// Cached implements the cache-aside (lazy-loading) pattern. It first
// attempts to read from cache; on a miss, it calls the loader function,
// caches the result, and returns it. A nil result from the loader is
// cached with a short TTL to prevent penetration attacks.
func Cached[T any](c Cache, key string, ttl time.Duration, loader func() (T, error)) (T, error) {
	var result T
	err := c.Get(context.Background(), key, &result)
	if err == nil {
		return result, nil
	}

	if err != redis.Nil {
		// Non-miss error; fall through to loader.
	}

	result, err = loader()
	if err != nil {
		return result, err
	}

	// Cache the result. Use a shorter TTL for empty values to prevent
	// cache penetration while allowing quick recovery.
	cacheTTL := ttl
	var zero T
	if any(result) == any(zero) {
		cacheTTL = 5 * time.Minute
	}

	_ = c.Set(context.Background(), key, result, cacheTTL)
	return result, nil
}
