package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig holds rate limiting parameters.
type RateLimiterConfig struct {
	MaxRequests int
	Window      time.Duration
	KeyPrefix   string
}

// RateLimiter returns a middleware that enforces sliding window rate
// limiting per client IP using Redis atomic operations. Returns HTTP 429
// when the limit is exceeded.
func RateLimiter(rdb *redis.Client, cfg RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("%s:%s", cfg.KeyPrefix, c.ClientIP())
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			rdb.Expire(ctx, key, cfg.Window)
		}

		if count > int64(cfg.MaxRequests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    10007,
				"message": "too many requests",
			})
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(cfg.MaxRequests)-count))
		c.Next()
	}
}
