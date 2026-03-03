package httpclient

import (
	"net"
	"net/http"
	"time"
)

// PoolConfig holds configuration for the shared HTTP client pool.
type PoolConfig struct {
	MaxIdleConns        int           `mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `mapstructure:"max_idle_conns_per_host"`
	MaxConnsPerHost     int           `mapstructure:"max_conns_per_host"`
	IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout"`
	RequestTimeout      time.Duration `mapstructure:"request_timeout"`
}

// DefaultPoolConfig returns sensible defaults for high-concurrency outbound requests
// to many different hosts (e.g., hundreds of AstrBot instances).
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     90 * time.Second,
		RequestTimeout:      30 * time.Second,
	}
}

// NewPool creates a shared http.Client with a tuned transport for high-concurrency
// outbound requests. The client should be reused across all adapters to benefit from
// connection pooling and keep-alive.
func NewPool(cfg PoolConfig) *http.Client {
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 200
	}
	if cfg.MaxIdleConnsPerHost <= 0 {
		cfg.MaxIdleConnsPerHost = 10
	}
	if cfg.MaxConnsPerHost <= 0 {
		cfg.MaxConnsPerHost = 20
	}
	if cfg.IdleConnTimeout <= 0 {
		cfg.IdleConnTimeout = 90 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 30 * time.Second
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.RequestTimeout,
	}
}
