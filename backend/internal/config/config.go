package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Storage    StorageConfig    `mapstructure:"storage"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Queue      QueueConfig      `mapstructure:"queue"`
	Log        LogConfig        `mapstructure:"log"`
	Email      EmailConfig      `mapstructure:"email"`
	Moderation ModerationConfig `mapstructure:"moderation"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Mode           string   `mapstructure:"mode"`
	Port           int      `mapstructure:"port"`
	ReadTimeout    int      `mapstructure:"read_timeout"`
	WriteTimeout   int      `mapstructure:"write_timeout"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	FrontendURL    string   `mapstructure:"frontend_url"`
}

// DatabaseConfig holds PostgreSQL connection configuration.
// Supports both component-based DSN (host/port/user/password) for local databases
// and direct connection string (dsn_direct) for cloud databases (Alibaba Cloud RDS,
// Neon, Supabase, etc.).
type DatabaseConfig struct {
	DSNDirect       string `mapstructure:"dsn_direct"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// DSN returns the PostgreSQL connection string.
// If dsn_direct is set, it takes precedence (for cloud databases).
func (c *DatabaseConfig) DSN() string {
	if c.DSNDirect != "" {
		return c.DSNDirect
	}
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, sslMode,
	)
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// StorageConfig holds object storage configuration.
type StorageConfig struct {
	Provider  string          `mapstructure:"provider"`
	MinIO     MinIOConfig     `mapstructure:"minio"`
	AliyunOSS AliyunOSSConfig `mapstructure:"aliyun_oss"`
}

// MinIOConfig holds MinIO connection configuration.
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Bucket    string `mapstructure:"bucket"`
}

// AliyunOSSConfig holds Alibaba Cloud OSS configuration.
type AliyunOSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Bucket          string `mapstructure:"bucket"`
	CDNDomain       string `mapstructure:"cdn_domain"`
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpire  int    `mapstructure:"access_token_expire"`
	RefreshTokenExpire int    `mapstructure:"refresh_token_expire"`
	Issuer             string `mapstructure:"issuer"`
}

// AccessTokenDuration returns the access token expiration as time.Duration.
func (c *JWTConfig) AccessTokenDuration() time.Duration {
	return time.Duration(c.AccessTokenExpire) * time.Minute
}

// RefreshTokenDuration returns the refresh token expiration as time.Duration.
func (c *JWTConfig) RefreshTokenDuration() time.Duration {
	return time.Duration(c.RefreshTokenExpire) * time.Minute
}

// QueueConfig holds async task queue configuration.
type QueueConfig struct {
	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	Concurrency   int    `mapstructure:"concurrency"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// ModerationConfig holds content moderation configuration.
type ModerationConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`
}

// EmailConfig holds SMTP email configuration (supports Alibaba Cloud DirectMail).
type EmailConfig struct {
	SMTPHost    string `mapstructure:"smtp_host"`
	SMTPPort    int    `mapstructure:"smtp_port"`
	FromAddress string `mapstructure:"from_address"`
	FromName    string `mapstructure:"from_name"`
	Password    string `mapstructure:"password"`
	UseTLS      bool   `mapstructure:"use_tls"`
}

// Load reads configuration from file and environment variables.
// It loads .env files first (if present), then reads the config file,
// with environment variables taking highest precedence.
func Load(configPath string) (*Config, error) {
	// Load .env file if present (ignores error if file doesn't exist).
	_ = godotenv.Load()
	// Also try environment-specific .env files.
	_ = godotenv.Load(".env.local")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
