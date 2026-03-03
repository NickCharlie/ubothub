package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/NickCharlie/ubothub/backend/internal/adapter"
	"github.com/NickCharlie/ubothub/backend/internal/captcha"
	"github.com/NickCharlie/ubothub/backend/internal/config"
	"github.com/NickCharlie/ubothub/backend/internal/event"
	"github.com/NickCharlie/ubothub/backend/internal/handler"
	"github.com/NickCharlie/ubothub/backend/internal/middleware"
	"github.com/NickCharlie/ubothub/backend/internal/moderation"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/internal/storage"
	"github.com/NickCharlie/ubothub/backend/pkg/logger"
	"github.com/NickCharlie/ubothub/backend/pkg/token"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup creates and configures the Gin router with all routes and middleware.
func Setup(db *gorm.DB, rdb *redis.Client, store storage.ObjectStorage, cfg *config.Config, rootLogger *zap.Logger) *gin.Engine {
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

	// Determine storage bucket.
	bucket := cfg.Storage.MinIO.Bucket
	if cfg.Storage.Provider == "aliyun_oss" {
		bucket = cfg.Storage.AliyunOSS.Bucket
	}

	// Initialize event bus.
	eventLog := logger.Named(rootLogger, "event")
	eventBus := event.NewBus(10, eventLog)

	// Initialize adapter factory.
	adapterFactory := adapter.NewFactory()

	// Initialize repositories.
	userRepo := repository.NewUserRepository(db)
	botRepo := repository.NewBotRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	avatarRepo := repository.NewAvatarRepository(db)
	legalRepo := repository.NewLegalRepository(db)

	// Initialize content moderation service.
	moderationLog := logger.Named(rootLogger, "moderation")
	moderator := moderation.NewAliyunService(moderation.AliyunConfig{
		Enabled:         cfg.Moderation.Enabled,
		AccessKeyID:     cfg.Moderation.AccessKeyID,
		AccessKeySecret: cfg.Moderation.AccessKeySecret,
		Endpoint:        cfg.Moderation.Endpoint,
	}, moderationLog)

	// Initialize services.
	authLog := logger.Named(rootLogger, "auth")
	userLog := logger.Named(rootLogger, "handler")
	botLog := logger.Named(rootLogger, "bot")
	assetLog := logger.Named(rootLogger, "asset")
	authSvc := service.NewAuthService(userRepo, tokenMgr, rdb, authLog)
	userSvc := service.NewUserService(userRepo, userLog)
	botSvc := service.NewBotService(botRepo, botLog)
	assetSvc := service.NewAssetService(assetRepo, store, bucket, assetLog)
	avatarLog := logger.Named(rootLogger, "avatar")
	avatarSvc := service.NewAvatarService(avatarRepo, botRepo, avatarLog)
	legalLog := logger.Named(rootLogger, "legal")
	legalSvc := service.NewLegalService(legalRepo, legalLog)

	// Initialize captcha service (Redis-backed for distributed deployments).
	captchaSvc := captcha.NewService(rdb)

	// Initialize handlers.
	authHandler := handler.NewAuthHandler(authSvc, legalSvc, captchaSvc)
	userHandler := handler.NewUserHandler(userSvc)
	botHandler := handler.NewBotHandler(botSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)
	avatarHandler := handler.NewAvatarHandler(avatarSvc)
	legalHandler := handler.NewLegalHandler(legalSvc)
	gatewayLog := logger.Named(rootLogger, "gateway")
	gatewayHandler := handler.NewGatewayHandler(botSvc, adapterFactory, eventBus, moderator, gatewayLog)

	// Public auth routes with stricter rate limiting.
	authRoutes := r.Group("/api/v1/auth")
	authRoutes.Use(middleware.RateLimiter(rdb, middleware.RateLimiterConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		KeyPrefix:   "rl:auth",
	}))
	{
		authRoutes.GET("/captcha", authHandler.Captcha)
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/refresh", authHandler.Refresh)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	// Public asset browsing (no auth required).
	r.GET("/api/v1/assets/public", assetHandler.ListPublic)

	// Public legal agreement routes (no auth required).
	r.GET("/api/v1/legal/terms", legalHandler.GetTermsOfService)
	r.GET("/api/v1/legal/privacy", legalHandler.GetPrivacyPolicy)
	r.GET("/api/v1/legal/agreements", legalHandler.GetAllAgreements)

	// Public bot plaza and avatar preview for guest users (view only, no chat).
	r.GET("/api/v1/plaza/bots", botHandler.ListPublic)
	r.GET("/api/v1/plaza/bots/:id", botHandler.GetPublic)
	r.GET("/api/v1/plaza/avatars/:id", avatarHandler.GetPublic)

	// Bot gateway routes (authenticated by bot access token, not JWT).
	gatewayRoutes := r.Group("/api/v1/gateway")
	gatewayRoutes.Use(middleware.RateLimiter(rdb, middleware.RateLimiterConfig{
		MaxRequests: 300,
		Window:      time.Minute,
		KeyPrefix:   "rl:gateway",
	}))
	{
		gatewayRoutes.POST("/webhook/:token", gatewayHandler.Webhook)
		gatewayRoutes.POST("/message", gatewayHandler.Message)
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

		// Asset management routes.
		authGroup.GET("/assets", assetHandler.List)
		authGroup.POST("/assets/upload/presigned", assetHandler.PresignedUpload)
		authGroup.POST("/assets/upload/complete", assetHandler.CompleteUpload)
		authGroup.GET("/assets/:id", assetHandler.Get)
		authGroup.PUT("/assets/:id", assetHandler.Update)
		authGroup.DELETE("/assets/:id", assetHandler.Delete)
		authGroup.GET("/assets/:id/download", assetHandler.Download)
		authGroup.GET("/assets/:id/thumbnail", assetHandler.Thumbnail)

		// Avatar management routes.
		authGroup.GET("/avatars", avatarHandler.List)
		authGroup.POST("/avatars", avatarHandler.Create)
		authGroup.GET("/avatars/:id", avatarHandler.Get)
		authGroup.PUT("/avatars/:id", avatarHandler.Update)
		authGroup.DELETE("/avatars/:id", avatarHandler.Delete)
		authGroup.POST("/avatars/:id/bind-bot", avatarHandler.BindBot)
		authGroup.POST("/avatars/:id/bind-asset", avatarHandler.BindAsset)
		authGroup.DELETE("/avatars/:id/assets/:assetId", avatarHandler.UnbindAsset)
		authGroup.PUT("/avatars/:id/action-mapping", avatarHandler.UpdateActionMapping)
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
