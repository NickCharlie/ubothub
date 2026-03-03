package moderation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Result represents the outcome of a content moderation check.
type Result struct {
	Pass     bool              `json:"pass"`
	Labels   []string          `json:"labels,omitempty"`
	Reason   string            `json:"reason,omitempty"`
	RiskRate map[string]string `json:"risk_rate,omitempty"`
}

// Service defines the content moderation interface.
type Service interface {
	CheckText(ctx context.Context, text string) (*Result, error)
}

// AliyunConfig holds Alibaba Cloud Content Safety configuration.
type AliyunConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`
	Enabled         bool   `mapstructure:"enabled"`
}

// aliyunService implements content moderation using Alibaba Cloud Green (Content Safety).
type aliyunService struct {
	config AliyunConfig
	client *http.Client
	logger *zap.Logger
}

// NewAliyunService creates a new Alibaba Cloud content moderation service.
func NewAliyunService(cfg AliyunConfig, logger *zap.Logger) Service {
	if !cfg.Enabled {
		return &noopService{}
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://green-cip.cn-shanghai.aliyuncs.com"
	}
	return &aliyunService{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

// aliyunTextRequest represents the Alibaba Cloud text moderation API request.
type aliyunTextRequest struct {
	Service    string                 `json:"Service"`
	ServiceParameters string          `json:"ServiceParameters"`
}

type aliyunServiceParams struct {
	Content string `json:"content"`
}

// aliyunTextResponse represents the Alibaba Cloud text moderation API response.
type aliyunTextResponse struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    struct {
		Labels  string `json:"labels"`
		Reason  string `json:"reason"`
	} `json:"Data"`
	RequestID string `json:"RequestId"`
}

func (s *aliyunService) CheckText(ctx context.Context, text string) (*Result, error) {
	params, _ := json.Marshal(aliyunServiceParams{Content: text})
	reqBody, _ := json.Marshal(aliyunTextRequest{
		Service:           "chat_detection",
		ServiceParameters: string(params),
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.config.Endpoint+"/api/v1/moderation/text/scan",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-action", "TextModeration")
	req.Header.Set("x-acs-version", "2022-03-02")

	// Sign with Alibaba Cloud credentials (simplified for now, production should use official SDK).
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.AccessKeyID))

	httpResp, err := s.client.Do(req)
	if err != nil {
		s.logger.Warn("content moderation request failed, allowing content",
			zap.Error(err),
		)
		// Fail open: if moderation service is unavailable, allow content through.
		return &Result{Pass: true}, nil
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &Result{Pass: true}, nil
	}

	var resp aliyunTextResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		s.logger.Warn("failed to parse moderation response",
			zap.Error(err),
			zap.String("body", string(body)),
		)
		return &Result{Pass: true}, nil
	}

	if resp.Code != 200 {
		s.logger.Warn("moderation service returned error",
			zap.Int("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return &Result{Pass: true}, nil
	}

	// Empty labels means content is clean.
	if resp.Data.Labels == "" {
		return &Result{Pass: true}, nil
	}

	s.logger.Info("content flagged by moderation",
		zap.String("labels", resp.Data.Labels),
		zap.String("reason", resp.Data.Reason),
	)

	return &Result{
		Pass:   false,
		Labels: []string{resp.Data.Labels},
		Reason: resp.Data.Reason,
	}, nil
}

// noopService is a no-op moderation service used when moderation is disabled.
type noopService struct{}

func (s *noopService) CheckText(_ context.Context, _ string) (*Result, error) {
	return &Result{Pass: true}, nil
}
