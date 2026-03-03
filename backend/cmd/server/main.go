package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NickCharlie/ubothub/backend/internal/config"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/queue"
	"github.com/NickCharlie/ubothub/backend/internal/queue/tasks"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/internal/router"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/internal/storage"
	"github.com/NickCharlie/ubothub/backend/pkg/email"
	"github.com/NickCharlie/ubothub/backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var mode = flag.String("mode", "all", "Run mode: api, worker, all")

func main() {
	flag.Parse()

	// Load configuration.
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger with datetime, module, priority, weight.
	rootLogger := logger.New(logger.Config{
		Level:    cfg.Log.Level,
		Format:   cfg.Log.Format,
		Output:   cfg.Log.Output,
		FilePath: cfg.Log.FilePath,
	})
	defer rootLogger.Sync()

	serverLog := logger.Named(rootLogger, "server")
	serverLog.Info("starting ubothub server", zap.String("mode", *mode))

	// Initialize database.
	dbLog := logger.Named(rootLogger, "database")
	db, err := initDatabase(cfg.Database)
	if err != nil {
		dbLog.Fatal("failed to connect database", zap.Error(err))
	}
	dbLog.Info("database connected",
		zap.String("host", cfg.Database.Host),
		zap.Bool("cloud", cfg.Database.DSNDirect != ""),
	)

	// Auto-migrate models.
	if err := autoMigrate(db); err != nil {
		dbLog.Fatal("failed to auto-migrate", zap.Error(err))
	}
	dbLog.Info("database migration completed")

	// Seed default legal agreements.
	legalLog := logger.Named(rootLogger, "legal")
	legalRepo := repository.NewLegalRepository(db)
	legalSvc := service.NewLegalService(legalRepo, legalLog)
	if err := legalSvc.SeedDefaultAgreements(context.Background()); err != nil {
		legalLog.Warn("failed to seed default agreements", zap.Error(err))
	}

	// Initialize Redis.
	redisLog := logger.Named(rootLogger, "redis")
	rdb := initRedis(cfg.Redis)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		redisLog.Fatal("failed to connect redis", zap.Error(err))
	}
	redisLog.Info("redis connected")

	// Initialize object storage.
	storageLog := logger.Named(rootLogger, "storage")
	store, err := storage.NewFromConfig(cfg.Storage)
	if err != nil {
		storageLog.Fatal("failed to initialize storage", zap.Error(err))
	}
	storageLog.Info("object storage initialized", zap.String("provider", cfg.Storage.Provider))

	// Initialize async task queue client.
	queueLog := logger.Named(rootLogger, "queue")
	queueClient := queue.NewClient(cfg.Queue.RedisAddr, cfg.Queue.RedisPassword, queueLog)
	defer queueClient.Close()

	// Start async worker if mode is "worker" or "all".
	var queueServer *queue.Server
	if *mode == "worker" || *mode == "all" {
		queueServer = queue.NewServer(
			cfg.Queue.RedisAddr,
			cfg.Queue.RedisPassword,
			cfg.Queue.Concurrency,
			queueLog,
		)

		// Register task handlers.
		emailSender := email.NewSender(email.Config{
			SMTPHost:    cfg.Email.SMTPHost,
			SMTPPort:    cfg.Email.SMTPPort,
			FromAddress: cfg.Email.FromAddress,
			FromName:    cfg.Email.FromName,
			Password:    cfg.Email.Password,
			UseTLS:      cfg.Email.UseTLS,
		})
		emailHandler := queue.NewEmailHandler(emailSender, logger.Named(rootLogger, "queue.email"))
		queueServer.Register(tasks.TypeEmailSend, emailHandler)

		msgLogRepo := repository.NewMessageLogRepository(db)
		msgLogHandler := queue.NewMessageLogHandler(msgLogRepo, logger.Named(rootLogger, "queue.message_log"))
		queueServer.Register(tasks.TypeMessageLog, msgLogHandler)

		go func() {
			if err := queueServer.Start(); err != nil {
				queueLog.Fatal("failed to start async worker", zap.Error(err))
			}
		}()
	}

	// Start HTTP server if mode is "api" or "all".
	var srv *http.Server
	if *mode == "api" || *mode == "all" {
		gin.SetMode(cfg.Server.Mode)

		r := router.Setup(db, rdb, store, cfg, rootLogger)

		srv = &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      r,
			ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		}

		go func() {
			serverLog.Info("HTTP server listening", zap.Int("port", cfg.Server.Port))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverLog.Fatal("server error", zap.Error(err))
			}
		}()
	}

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	serverLog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			serverLog.Fatal("server forced to shutdown", zap.Error(err))
		}
	}

	if queueServer != nil {
		queueServer.Stop()
	}

	serverLog.Info("server exited gracefully")
}

// initDatabase establishes a PostgreSQL connection via GORM.
// Supports both DSN-component mode and direct connection string for cloud databases.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

// autoMigrate runs GORM auto-migration for all registered models.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Bot{},
		&model.Asset{},
		&model.AvatarConfig{},
		&model.AvatarAsset{},
		&model.MessageLog{},
		&model.ActionTemplate{},
		&model.LegalAgreement{},
		&model.UserAgreementAcceptance{},
	)
}

// initRedis creates a Redis client.
func initRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
}
