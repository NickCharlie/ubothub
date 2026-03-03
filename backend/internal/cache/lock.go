package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedLock provides a simple Redis-based distributed lock to
// coordinate exclusive access across multiple processes or workers.
type DistributedLock struct {
	client *redis.Client
}

// NewDistributedLock creates a new distributed lock backed by Redis.
func NewDistributedLock(client *redis.Client) *DistributedLock {
	return &DistributedLock{client: client}
}

// Acquire attempts to obtain the lock identified by key with the given TTL.
// Returns true if the lock was successfully acquired.
func (l *DistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := l.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Release removes the lock identified by key, allowing other callers
// to acquire it.
func (l *DistributedLock) Release(ctx context.Context, key string) error {
	return l.client.Del(ctx, "lock:"+key).Err()
}
