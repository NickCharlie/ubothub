package captcha

import (
	"context"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

// Service wraps base64Captcha with a Redis-backed store for distributed deployments.
type Service struct {
	captcha *base64Captcha.Captcha
}

// GenerateResult holds the generated captcha ID and base64 image.
type GenerateResult struct {
	CaptchaID string `json:"captcha_id"`
	Image     string `json:"captcha_image"`
}

// NewService creates a new captcha service backed by Redis.
// Uses math driver (arithmetic like "3 + 7 = ?") for better UX.
func NewService(rdb *redis.Client) *Service {
	store := &redisStore{rdb: rdb, expiration: 5 * time.Minute}

	driver := base64Captcha.NewDriverMath(
		60,  // height
		240, // width
		5,   // noiseCount
		base64Captcha.OptionShowSlimeLine,
		nil, // background color (default)
		nil, // fonts (default built-in)
	)

	c := base64Captcha.NewCaptcha(driver, store)
	return &Service{captcha: c}
}

// Generate creates a new captcha and returns its ID + base64 image string.
func (s *Service) Generate() (*GenerateResult, error) {
	id, b64s, err := s.captcha.Generate()
	if err != nil {
		return nil, err
	}
	return &GenerateResult{
		CaptchaID: id,
		Image:     b64s,
	}, nil
}

// Verify checks whether the user's answer matches the stored captcha.
// The captcha is consumed after verification (one-time use).
func (s *Service) Verify(captchaID, answer string) bool {
	return s.captcha.Verify(captchaID, answer, true)
}

// redisStore implements base64Captcha.Store using Redis for distributed deployments.
type redisStore struct {
	rdb        *redis.Client
	expiration time.Duration
}

// Set stores the captcha answer in Redis with TTL.
func (s *redisStore) Set(id string, value string) {
	s.rdb.Set(context.Background(), captchaKey(id), value, s.expiration)
}

// Get retrieves the captcha answer from Redis. If clear is true, deletes the key.
func (s *redisStore) Get(id string, clear bool) string {
	key := captchaKey(id)
	val, err := s.rdb.Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	if clear {
		s.rdb.Del(context.Background(), key)
	}
	return val
}

// Verify checks the answer against the stored value.
func (s *redisStore) Verify(id, answer string, clear bool) bool {
	stored := s.Get(id, clear)
	return stored != "" && stored == answer
}

func captchaKey(id string) string {
	return "captcha:" + id
}
