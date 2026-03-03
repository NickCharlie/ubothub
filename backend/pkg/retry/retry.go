package retry

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// Config holds retry configuration.
type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultConfig returns sensible defaults for external API retries.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

// Do executes fn with exponential backoff and jitter. It retries on error
// up to MaxAttempts times, respecting context cancellation.
func Do(ctx context.Context, cfg Config, fn func(ctx context.Context) error) error {
	var lastErr error

	for attempt := range cfg.MaxAttempts {
		if err := ctx.Err(); err != nil {
			return err
		}

		if lastErr = fn(ctx); lastErr == nil {
			return nil
		}

		// Don't sleep after the last attempt.
		if attempt < cfg.MaxAttempts-1 {
			delay := backoffWithJitter(cfg.BaseDelay, cfg.MaxDelay, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return lastErr
}

// backoffWithJitter calculates exponential backoff with full jitter.
func backoffWithJitter(base, max time.Duration, attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(base) * exp)
	if delay > max {
		delay = max
	}
	// Full jitter: random value in [0, delay].
	return time.Duration(rand.Int64N(int64(delay) + 1))
}
