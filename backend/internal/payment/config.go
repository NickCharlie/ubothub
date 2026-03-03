package payment

// Channel constants identify the payment method used.
const (
	ChannelWechatNative = "wechat_native" // WeChat QR code payment
	ChannelWechatJSAPI  = "wechat_jsapi"  // WeChat in-app payment (mini-program/official account)
	ChannelWechatH5     = "wechat_h5"     // WeChat mobile browser payment
	ChannelAlipayPage   = "alipay_page"   // Alipay PC page payment
	ChannelAlipayWap    = "alipay_wap"    // Alipay mobile WAP payment
	ChannelAlipayQR     = "alipay_qr"     // Alipay QR code payment (precreate)
)

// Order status constants.
const (
	OrderStatusPending  = "pending"
	OrderStatusPaid     = "paid"
	OrderStatusClosed   = "closed"
	OrderStatusRefunded = "refunded"
	OrderStatusFailed   = "failed"
)

// WechatConfig holds WeChat Pay V3 service provider configuration.
type WechatConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	SpMchID    string `mapstructure:"sp_mch_id"`    // service provider merchant ID
	SpAppID    string `mapstructure:"sp_app_id"`    // service provider app ID
	SubMchID   string `mapstructure:"sub_mch_id"`   // sub-merchant ID (platform merchant)
	SubAppID   string `mapstructure:"sub_app_id"`   // sub-merchant app ID (optional)
	SerialNo   string `mapstructure:"serial_no"`    // API certificate serial number
	ApiV3Key   string `mapstructure:"api_v3_key"`   // APIv3 key
	PrivateKey string `mapstructure:"private_key"`  // merchant API certificate private key (PEM)
	NotifyURL  string `mapstructure:"notify_url"`   // payment notification callback URL
}

// AlipayConfig holds Alipay service provider configuration.
type AlipayConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	AppID           string `mapstructure:"app_id"`            // ISV app ID
	PrivateKey      string `mapstructure:"private_key"`       // app private key (RSA2, PKCS1)
	AlipayPublicKey string `mapstructure:"alipay_public_key"` // alipay public key for verification
	IsProd          bool   `mapstructure:"is_prod"`           // true for production, false for sandbox
	NotifyURL       string `mapstructure:"notify_url"`        // payment notification callback URL
	ReturnURL       string `mapstructure:"return_url"`        // payment return URL after completion
}
