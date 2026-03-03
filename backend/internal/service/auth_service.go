package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/repository"
	"github.com/NickCharlie/ubothub/backend/pkg/hash"
	"github.com/NickCharlie/ubothub/backend/pkg/token"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthService handles user authentication business logic.
type AuthService struct {
	userRepo repository.UserRepository
	tokenMgr *token.Manager
	rdb      *redis.Client
	logger   *zap.Logger
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	userRepo repository.UserRepository,
	tokenMgr *token.Manager,
	rdb *redis.Client,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		tokenMgr: tokenMgr,
		rdb:      rdb,
		logger:   logger,
	}
}

// RegisterResult contains the result of a successful registration.
type RegisterResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

// Register creates a new user account after validation.
func (s *AuthService) Register(ctx context.Context, email, username, password string) (*RegisterResult, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email existence: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("check username existence: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		ID:           xid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  username,
		Role:         "user",
		Status:       "active",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	accessToken, err := s.tokenMgr.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenMgr.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	s.logger.Info("user registered", zap.String("user_id", user.ID), zap.String("email", email))

	return &RegisterResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

// Login authenticates a user by email and password.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	// Check login failure rate limit.
	failKey := fmt.Sprintf("login:fail:%s", email)
	failCount, _ := s.rdb.Get(ctx, failKey).Int64()
	if failCount >= 5 {
		return nil, ErrAccountLocked
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.incrementLoginFailure(ctx, failKey)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if user.Status != "active" {
		return nil, ErrAccountLocked
	}

	if !hash.CheckPassword(password, user.PasswordHash) {
		s.incrementLoginFailure(ctx, failKey)
		return nil, ErrInvalidCredentials
	}

	// Clear failure count on successful login.
	s.rdb.Del(ctx, failKey)

	accessToken, err := s.tokenMgr.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.tokenMgr.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	s.logger.Info("user logged in", zap.String("user_id", user.ID))

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken validates a refresh token and issues new token pair.
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	claims, err := s.tokenMgr.ParseToken(refreshTokenStr)
	if err != nil {
		return "", "", ErrTokenInvalid
	}

	// Check if refresh token is blacklisted.
	blacklistKey := "jwt:blacklist:" + claims.ID
	exists, _ := s.rdb.Exists(ctx, blacklistKey).Result()
	if exists > 0 {
		return "", "", ErrTokenRevoked
	}

	// Blacklist the old refresh token.
	if claims.ExpiresAt != nil {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			s.rdb.Set(ctx, blacklistKey, "1", ttl)
		}
	}

	accessToken, err := s.tokenMgr.GenerateAccessToken(claims.UserID, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenMgr.GenerateRefreshToken(claims.UserID, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// Logout invalidates the given access token by adding it to the blacklist.
func (s *AuthService) Logout(ctx context.Context, tokenStr string) error {
	claims, err := s.tokenMgr.ParseToken(tokenStr)
	if err != nil {
		return nil
	}

	if claims.ExpiresAt != nil {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			s.rdb.Set(ctx, "jwt:blacklist:"+claims.ID, "1", ttl)
		}
	}

	s.logger.Info("user logged out", zap.String("user_id", claims.UserID))
	return nil
}

func (s *AuthService) incrementLoginFailure(ctx context.Context, key string) {
	pipe := s.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 15*time.Minute)
	pipe.Exec(ctx)
}
