package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ubothub/backend/internal/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup creates and configures the Gin router with all routes and middleware.
func Setup(db *gorm.DB, rdb *redis.Client, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	r := gin.New()

	// Global middleware will be added in subsequent phases:
	// - Recovery
	// - RequestID
	// - Logger
	// - CORS
	// - SecurityHeaders
	// - RateLimiter

	// Health check endpoint.
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "ubothub",
		})
	})

	// API v1 route group.
	// Route registration for each module will be added in subsequent phases.
	_ = r.Group("/api/v1")

	return r
}
