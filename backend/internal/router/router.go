package router

import (
	"context"
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
	"github.com/NickCharlie/ubothub/backend/internal/payment"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/internal/storage"
	"github.com/NickCharlie/ubothub/backend/internal/ws"
	"github.com/NickCharlie/ubothub/backend/pkg/email"
	"github.com/NickCharlie/ubothub/backend/pkg/httpclient"
	"github.com/NickCharlie/ubothub/backend/pkg/logger"
	"github.com/NickCharlie/ubothub/backend/pkg/token"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup creates and configures the Gin router with all routes and middleware.
// The returned context cancel function should be called during shutdown to
// gracefully close the WebSocket hub.
func Setup(ctx context.Context, db *gorm.DB, rdb *redis.Client, store storage.ObjectStorage, cfg *config.Config, rootLogger *zap.Logger) *gin.Engine {
	r := gin.New()

	mwLog := logger.Named(rootLogger, "middleware")

	// Global middleware chain (onion model, executed in order).
	r.Use(middleware.Recovery(mwLog))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(mwLog))
	r.Use(middleware.CORS(cfg.Server.Mode, cfg.Server.AllowedOrigins))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CSRF())
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

	// Initialize shared HTTP client pool for outbound requests to AstrBot instances.
	httpClient := httpclient.NewPool(httpclient.DefaultPoolConfig())

	// Initialize adapter factory with shared HTTP client.
	adapterFactory := adapter.NewFactory(httpClient)

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

	// Initialize email service (Alibaba Cloud DirectMail via SMTP).
	emailLog := logger.Named(rootLogger, "email")
	emailSender := email.NewSender(email.Config{
		SMTPHost:    cfg.Email.SMTPHost,
		SMTPPort:    cfg.Email.SMTPPort,
		FromAddress: cfg.Email.FromAddress,
		FromName:    cfg.Email.FromName,
		Password:    cfg.Email.Password,
		UseTLS:      cfg.Email.UseTLS,
	})
	emailSvc := service.NewEmailService(emailSender, rdb, cfg.Server.FrontendURL, emailLog)

	// Initialize handlers.
	authHandler := handler.NewAuthHandler(authSvc, emailSvc, legalSvc, captchaSvc)
	userHandler := handler.NewUserHandler(userSvc)
	botHandler := handler.NewBotHandler(botSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)
	avatarHandler := handler.NewAvatarHandler(avatarSvc)
	legalHandler := handler.NewLegalHandler(legalSvc)
	gatewayLog := logger.Named(rootLogger, "gateway")
	gatewayHandler := handler.NewGatewayHandler(botSvc, adapterFactory, eventBus, moderator, gatewayLog)

	// Initialize admin service and handler.
	adminRepo := repository.NewAdminRepository(db)
	adminLog := logger.Named(rootLogger, "admin")
	adminSvc := service.NewAdminService(adminRepo, adminLog)
	adminHandler := handler.NewAdminHandler(adminSvc)

	// Initialize wallet, billing, and payment services.
	walletRepo := repository.NewWalletRepository(db)
	txnRepo := repository.NewTransactionRepository(db)
	pricingRepo := repository.NewBotPricingRepository(db)
	walletLog := logger.Named(rootLogger, "wallet")
	walletSvc := service.NewWalletService(walletRepo, txnRepo, db, walletLog)
	billingLog := logger.Named(rootLogger, "billing")
	billingSvc := service.NewBillingService(pricingRepo, walletSvc, botRepo, billingLog)

	// Initialize payment provider registry (WeChat Pay V3 + Alipay service provider mode).
	paymentLog := logger.Named(rootLogger, "payment")
	paymentRegistry, err := payment.NewRegistry(
		payment.WechatConfig{
			Enabled:    cfg.Payment.Wechat.Enabled,
			SpMchID:    cfg.Payment.Wechat.SpMchID,
			SpAppID:    cfg.Payment.Wechat.SpAppID,
			SubMchID:   cfg.Payment.Wechat.SubMchID,
			SubAppID:   cfg.Payment.Wechat.SubAppID,
			SerialNo:   cfg.Payment.Wechat.SerialNo,
			ApiV3Key:   cfg.Payment.Wechat.ApiV3Key,
			PrivateKey: cfg.Payment.Wechat.PrivateKey,
			NotifyURL:  cfg.Payment.Wechat.NotifyURL,
		},
		payment.AlipayConfig{
			Enabled:         cfg.Payment.Alipay.Enabled,
			AppID:           cfg.Payment.Alipay.AppID,
			PrivateKey:      cfg.Payment.Alipay.PrivateKey,
			AlipayPublicKey: cfg.Payment.Alipay.AlipayPublicKey,
			IsProd:          cfg.Payment.Alipay.IsProd,
			NotifyURL:       cfg.Payment.Alipay.NotifyURL,
			ReturnURL:       cfg.Payment.Alipay.ReturnURL,
		},
		paymentLog,
	)
	if err != nil {
		paymentLog.Fatal("failed to initialize payment registry", zap.Error(err))
	}

	// Get default provider for wallet handler (falls back to noop if none enabled).
	defaultPaymentPvd, _ := payment.NewProvider(
		payment.WechatConfig{Enabled: cfg.Payment.Wechat.Enabled, SpMchID: cfg.Payment.Wechat.SpMchID, SpAppID: cfg.Payment.Wechat.SpAppID, SubMchID: cfg.Payment.Wechat.SubMchID, SubAppID: cfg.Payment.Wechat.SubAppID, SerialNo: cfg.Payment.Wechat.SerialNo, ApiV3Key: cfg.Payment.Wechat.ApiV3Key, PrivateKey: cfg.Payment.Wechat.PrivateKey, NotifyURL: cfg.Payment.Wechat.NotifyURL},
		payment.AlipayConfig{Enabled: cfg.Payment.Alipay.Enabled, AppID: cfg.Payment.Alipay.AppID, PrivateKey: cfg.Payment.Alipay.PrivateKey, AlipayPublicKey: cfg.Payment.Alipay.AlipayPublicKey, IsProd: cfg.Payment.Alipay.IsProd, NotifyURL: cfg.Payment.Alipay.NotifyURL, ReturnURL: cfg.Payment.Alipay.ReturnURL},
		paymentLog,
	)
	walletHandler := handler.NewWalletHandler(walletSvc, billingSvc, defaultPaymentPvd)
	paymentHandler := handler.NewPaymentHandler(paymentRegistry, walletSvc, paymentLog)

	// Initialize WebSocket hub with configurable limits.
	wsCfg := ws.DefaultConfig()
	if cfg.WebSocket.MaxConnections > 0 {
		wsCfg.MaxConnections = cfg.WebSocket.MaxConnections
	}
	if cfg.WebSocket.MaxConnectionsPerRoom > 0 {
		wsCfg.MaxConnectionsPerRoom = cfg.WebSocket.MaxConnectionsPerRoom
	}
	if cfg.WebSocket.MaxConnectionsPerUser > 0 {
		wsCfg.MaxConnectionsPerUser = cfg.WebSocket.MaxConnectionsPerUser
	}
	if cfg.WebSocket.ReadBufferSize > 0 {
		wsCfg.ReadBufferSize = cfg.WebSocket.ReadBufferSize
	}
	if cfg.WebSocket.WriteBufferSize > 0 {
		wsCfg.WriteBufferSize = cfg.WebSocket.WriteBufferSize
	}
	if cfg.WebSocket.MaxMessageSize > 0 {
		wsCfg.MaxMessageSize = int64(cfg.WebSocket.MaxMessageSize)
	}
	wsLog := logger.Named(rootLogger, "websocket")
	wsHub := ws.NewHub(wsCfg, wsLog)

	// Register event bus subscribers for real-time WebSocket delivery.
	ws.RegisterEventSubscribers(wsHub, eventBus, wsLog)

	// Start the WebSocket hub event loop (blocks until ctx is cancelled).
	go wsHub.Run(ctx)

	wsHandler := handler.NewWSHandler(wsHub, tokenMgr, cfg.Server.AllowedOrigins, wsLog)

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
		authRoutes.GET("/verify-email", authHandler.VerifyEmail)
		authRoutes.POST("/forgot-password", authHandler.ForgotPassword)
		authRoutes.POST("/reset-password", authHandler.ResetPassword)
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

	// WebSocket endpoint (JWT validated via query parameter, not middleware).
	r.GET("/api/v1/ws", wsHandler.Connect)

	// Payment notification callback routes (no auth, verified by provider signature).
	paymentRoutes := r.Group("/api/v1/payment/notify")
	{
		paymentRoutes.POST("/wechat", paymentHandler.WechatNotify)
		paymentRoutes.POST("/alipay", paymentHandler.AlipayNotify)
	}

	// Authenticated API v1 route group.
	authGroup := r.Group("/api/v1")
	authGroup.Use(middleware.JWTAuth(tokenMgr, rdb))
	{
		// User profile routes.
		authGroup.GET("/users/me", userHandler.GetMe)
		authGroup.PUT("/users/me", userHandler.UpdateMe)
		authGroup.PUT("/users/me/password", userHandler.ChangePassword)
		authGroup.POST("/auth/resend-verification", authHandler.ResendVerification)

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

		// Wallet and billing routes.
		authGroup.GET("/wallet", walletHandler.GetWallet)
		authGroup.POST("/wallet/topup", walletHandler.TopUp)
		authGroup.GET("/wallet/transactions", walletHandler.Transactions)

		// Bot pricing routes (creator configures monetization).
		authGroup.GET("/bots/:id/pricing", walletHandler.GetBotPricing)
		authGroup.PUT("/bots/:id/pricing", walletHandler.SetBotPricing)
	}

	// Admin-only API v1 route group.
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.JWTAuth(tokenMgr, rdb))
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		adminGroup.GET("/dashboard", adminHandler.Dashboard)
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.PUT("/users/:id/ban", adminHandler.BanUser)
		adminGroup.PUT("/users/:id/unban", adminHandler.UnbanUser)
		adminGroup.GET("/bots", adminHandler.ListBots)
		adminGroup.DELETE("/bots/:id", adminHandler.ForceDeleteBot)
		adminGroup.GET("/assets", adminHandler.ListAssets)
		adminGroup.GET("/logs", adminHandler.ListMessageLogs)
		adminGroup.GET("/ws/metrics", wsHandler.Metrics)
	}

	return r
}
