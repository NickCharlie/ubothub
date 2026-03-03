package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ubothub/backend/internal/config"
	"github.com/ubothub/backend/internal/handler"
	"github.com/ubothub/backend/internal/middleware"
	"github.com/ubothub/backend/internal/repository"
	"github.com/ubothub/backend/internal/service"
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

	// JWT token manager.
	tokenMgr := token.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration(),
		cfg.JWT.RefreshTokenDuration(),
		cfg.JWT.Issuer,
	)

	// Initialize repositories.
	userRepo := repository.NewUserRepository(db)
	botRepo := repository.NewBotRepository(db)

	// Initialize services.
	authLog := logger.Named(rootLogger, "auth")
	userLog := logger.Named(rootLogger, "handler")
	botLog := logger.Named(rootLogger, "bot")
	authSvc := service.NewAuthService(userRepo, tokenMgr, rdb, authLog)
	userSvc := service.NewUserService(userRepo, userLog)
	botSvc := service.NewBotService(botRepo, botLog)

	// Initialize handlers.
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	botHandler := handler.NewBotHandler(botSvc)

	// Public auth routes with stricter rate limiting.
	authRoutes := r.Group("/api/v1/auth")
	authRoutes.Use(middleware.RateLimiter(rdb, middleware.RateLimiterConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		KeyPrefix:   "rl:auth",
	}))
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.Refresh)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	// Authenticated API v1 route group.
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.JWTAuth(tokenMgr, rdb))
	{
		// User profile routes.
		authGroup.GET("/users/me", userHandler.GetMe)
		authGroup.PUT("/users/me", userHandler.UpdateMe)
		authGroup.PUT("/users/me/password", userHandler.ChangePassword)

		// Bot management routes.
		authGroup.GET("/bots", botHandler.List)
		authGroup.POST("/bots", botHandler.Create)
		authGroup.GET("/bots/:id", botHandler.Get)
		authGroup.PUT("/bots/:id", botHandler.Update)
		authGroup.DELETE("/bots/:id", botHandler.Delete)
		authGroup.POST("/bots/:id/regenerate-token", botHandler.RegenerateToken)
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
