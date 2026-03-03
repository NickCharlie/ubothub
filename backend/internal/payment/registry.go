package payment

import (
	"fmt"

	"go.uber.org/zap"
)

// NewProvider creates a payment provider based on the specified channel.
// Returns noop provider when neither wechat nor alipay is enabled.
func NewProvider(wechatCfg WechatConfig, alipayCfg AlipayConfig, logger *zap.Logger) (Provider, error) {
	if wechatCfg.Enabled {
		return NewWechatProvider(wechatCfg, logger)
	}
	if alipayCfg.Enabled {
		return NewAlipayProvider(alipayCfg, logger)
	}
	logger.Info("no payment provider enabled, using noop provider")
	return NewNoopProvider(), nil
}

// Registry manages multiple payment providers keyed by channel name.
type Registry struct {
	providers map[string]Provider
	logger    *zap.Logger
}

// NewRegistry creates a payment provider registry from configuration.
// It initializes all enabled providers and registers them by name.
func NewRegistry(wechatCfg WechatConfig, alipayCfg AlipayConfig, logger *zap.Logger) (*Registry, error) {
	reg := &Registry{
		providers: make(map[string]Provider),
		logger:    logger,
	}

	if wechatCfg.Enabled {
		wp, err := NewWechatProvider(wechatCfg, logger)
		if err != nil {
			return nil, fmt.Errorf("init wechat provider: %w", err)
		}
		reg.providers["wechat"] = wp
		logger.Info("wechat payment provider registered")
	}

	if alipayCfg.Enabled {
		ap, err := NewAlipayProvider(alipayCfg, logger)
		if err != nil {
			return nil, fmt.Errorf("init alipay provider: %w", err)
		}
		reg.providers["alipay"] = ap
		logger.Info("alipay payment provider registered")
	}

	// Always register noop as fallback.
	reg.providers["noop"] = NewNoopProvider()

	return reg, nil
}

// Get returns the provider for the given channel name.
func (r *Registry) Get(channel string) (Provider, error) {
	// Map channel to provider name.
	providerName := channelToProvider(channel)
	p, ok := r.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("payment provider not found: %s", providerName)
	}
	return p, nil
}

// GetByName returns the provider by its name.
func (r *Registry) GetByName(name string) (Provider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("payment provider not found: %s", name)
	}
	return p, nil
}

// channelToProvider maps a payment channel to its provider name.
func channelToProvider(channel string) string {
	switch channel {
	case ChannelWechatNative, ChannelWechatJSAPI, ChannelWechatH5:
		return "wechat"
	case ChannelAlipayPage, ChannelAlipayWap, ChannelAlipayQR:
		return "alipay"
	default:
		return "noop"
	}
}
