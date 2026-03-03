package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ubothub/backend/internal/config"
	"github.com/ubothub/backend/internal/middleware"
	"github.com/ubothub/backend/pkg/logger"
	"github.com/ubothub/backend/pkg/token"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup creates and configures the Gin router with all routes and middleware.
func Setup(db *gorm.DB, rdb *redis.Client, cfg *config.Config, rootLogger *zap.Logger) *gin.Engine {
	r := gin.New()

	mwLog := logger.Named(rootLogger, "middleware")

	// Global middleware chain (onion model, executed in order).
	r.Use(middleware.Recovery(mwLog))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(mwLog))
	r.Use(middleware.CORS(cfg.Server.Mode, nil))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimiter(rdb, middleware.RateLimiterConfig{
		MaxRequests: 100,
		Window:      time.Second,
		KeyPrefix:   "rl:global",
	}))

	// Health check endpoint (no auth required).
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "ubothub",
		})
	})

	// JWT token manager for authenticated routes.
	tokenMgr := token.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration(),
		cfg.JWT.RefreshTokenDuration(),
		cfg.JWT.Issuer,
	)

	// Public API v1 route group (auth endpoints).
	_ = r.Group("/api/v1/auth")
	// Auth route registration will be added in subsequent phases.

	// Authenticated API v1 route group.
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.JWTAuth(tokenMgr, rdb))
	{
		// Route registration for each module will be added in subsequent phases.
		_ = authGroup
	}

	// Admin-only API v1 route group.
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.JWTAuth(tokenMgr, rdb))
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		_ = adminGroup
	}

	return r
}
