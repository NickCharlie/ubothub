package payment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// wechatProvider implements Provider for WeChat Pay V3 service provider mode.
// In service provider mode, the platform acts as sp_mchid and manages
// sub-merchants (sub_mchid). Funds flow through official WeChat Pay channels,
// and profit sharing is handled via the WeChat profit sharing API.
type wechatProvider struct {
	client *wechat.ClientV3
	cfg    WechatConfig
	logger *zap.Logger
}

// NewWechatProvider creates a WeChat Pay V3 provider in service provider mode.
func NewWechatProvider(cfg WechatConfig, logger *zap.Logger) (Provider, error) {
	client, err := wechat.NewClientV3(cfg.SpMchID, cfg.SerialNo, cfg.ApiV3Key, cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("init wechat v3 client: %w", err)
	}

	// Enable automatic signature verification with certificate auto-refresh.
	if err := client.AutoVerifySign(); err != nil {
		logger.Warn("wechat auto verify sign failed, manual verification required", zap.Error(err))
	}

	return &wechatProvider{
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (p *wechatProvider) Name() string { return "wechat" }

// CreateOrder creates a payment order using WeChat Native (QR code) API
// in service provider mode (V3PartnerTransactionNative).
func (p *wechatProvider) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	// Convert amount from CNY (decimal) to cents (int).
	amountCents := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	bm := make(gopay.BodyMap)
	bm.Set("sp_appid", p.cfg.SpAppID).
		Set("sp_mchid", p.cfg.SpMchID).
		Set("sub_mchid", p.cfg.SubMchID).
		Set("description", req.Description).
		Set("out_trade_no", req.OrderID).
		Set("notify_url", p.resolveNotifyURL(req.NotifyURL)).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", amountCents).
				Set("currency", "CNY")
		})

	if p.cfg.SubAppID != "" {
		bm.Set("sub_appid", p.cfg.SubAppID)
	}

	// Attach user_id for identification in callback.
	bm.Set("attach", req.UserID)

	wxRsp, err := p.client.V3PartnerTransactionNative(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("wechat native order: %w", err)
	}
	if wxRsp.Code != wechat.Success {
		return nil, fmt.Errorf("wechat native order failed: code=%d, error=%s", wxRsp.Code, wxRsp.Error)
	}

	return &CreateOrderResponse{
		OrderID:   req.OrderID,
		QRCodeURL: wxRsp.Response.CodeUrl,
		Channel:   ChannelWechatNative,
	}, nil
}

// QueryOrder queries payment order status using the out_trade_no.
func (p *wechatProvider) QueryOrder(ctx context.Context, orderID string) (*QueryOrderResponse, error) {
	bm := make(gopay.BodyMap)
	bm.Set("sp_mchid", p.cfg.SpMchID).
		Set("sub_mchid", p.cfg.SubMchID)

	wxRsp, err := p.client.V3PartnerQueryOrder(ctx, wechat.OutTradeNo, orderID, bm)
	if err != nil {
		return nil, fmt.Errorf("wechat query order: %w", err)
	}
	if wxRsp.Code != wechat.Success {
		return nil, fmt.Errorf("wechat query order failed: code=%d, error=%s", wxRsp.Code, wxRsp.Error)
	}

	resp := wxRsp.Response
	status := mapWechatTradeState(resp.TradeState)
	amount := decimal.NewFromInt(int64(resp.Amount.Total)).Div(decimal.NewFromInt(100))

	return &QueryOrderResponse{
		OrderID:    resp.OutTradeNo,
		ExternalID: resp.TransactionId,
		Amount:     amount,
		Status:     status,
		Channel:    ChannelWechatNative,
		PaidAt:     resp.SuccessTime,
	}, nil
}

// Transfer is not directly supported in WeChat Pay service provider mode.
// Withdrawals should use the merchant transfer API (V3Transfer) separately.
func (p *wechatProvider) Transfer(_ context.Context, _ TransferRequest) (*TransferResponse, error) {
	return nil, fmt.Errorf("wechat service provider transfer not implemented, use dedicated withdrawal flow")
}

// ParseNotify parses and decrypts a WeChat Pay V3 async payment notification.
func (p *wechatProvider) ParseNotify(_ context.Context, r *http.Request) (*NotifyResult, error) {
	notifyReq, err := wechat.V3ParseNotify(r)
	if err != nil {
		return nil, fmt.Errorf("parse wechat notify: %w", err)
	}

	result, err := notifyReq.DecryptPartnerPayCipherText(string(p.client.ApiV3Key))
	if err != nil {
		return nil, fmt.Errorf("decrypt wechat notify: %w", err)
	}

	status := mapWechatTradeState(result.TradeState)
	amount := decimal.NewFromInt(int64(result.Amount.Total)).Div(decimal.NewFromInt(100))

	return &NotifyResult{
		OrderID:    result.OutTradeNo,
		ExternalID: result.TransactionId,
		Amount:     amount,
		Status:     status,
		Channel:    ChannelWechatNative,
	}, nil
}

// VerifyNotify verifies the WeChat Pay V3 notification signature.
func (p *wechatProvider) VerifyNotify(_ context.Context, r *http.Request) (bool, error) {
	notifyReq, err := wechat.V3ParseNotify(r)
	if err != nil {
		return false, fmt.Errorf("parse wechat notify for verify: %w", err)
	}

	// Verify using platform certificate map.
	certMap := p.client.WxPublicKeyMap()
	err = notifyReq.VerifySignByPKMap(certMap)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (p *wechatProvider) resolveNotifyURL(override string) string {
	if override != "" {
		return override
	}
	return p.cfg.NotifyURL
}

// mapWechatTradeState maps WeChat trade state to internal order status.
func mapWechatTradeState(state string) string {
	switch state {
	case "SUCCESS":
		return OrderStatusPaid
	case "NOTPAY", "USERPAYING":
		return OrderStatusPending
	case "CLOSED":
		return OrderStatusClosed
	case "REFUND":
		return OrderStatusRefunded
	case "PAYERROR", "REVOKED":
		return OrderStatusFailed
	default:
		return OrderStatusPending
	}
}
