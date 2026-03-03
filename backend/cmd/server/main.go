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

	"github.com/ubothub/backend/internal/config"
	"github.com/ubothub/backend/internal/model"
	"github.com/ubothub/backend/internal/router"
	"github.com/ubothub/backend/pkg/logger"

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
	dbLog.Info("database connected")

	// Auto-migrate models.
	if err := autoMigrate(db); err != nil {
		dbLog.Fatal("failed to auto-migrate", zap.Error(err))
	}
	dbLog.Info("database migration completed")

	// Initialize Redis.
	redisLog := logger.Named(rootLogger, "redis")
	rdb := initRedis(cfg.Redis)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		redisLog.Fatal("failed to connect redis", zap.Error(err))
	}
	redisLog.Info("redis connected")

	// Set Gin mode.
	gin.SetMode(cfg.Server.Mode)

	// Create router with dependencies.
	r := router.Setup(db, rdb, cfg, rootLogger)

	// Start HTTP server.
	srv := &http.Server{
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

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	serverLog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		serverLog.Fatal("server forced to shutdown", zap.Error(err))
	}

	serverLog.Info("server exited gracefully")
}

// initDatabase establishes a PostgreSQL connection via GORM.
func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
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
