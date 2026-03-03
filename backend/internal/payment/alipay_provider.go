package payment

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// alipayProvider implements Provider for Alipay service provider mode.
// In service provider mode, the platform acts as ISV and processes payments
// on behalf of merchants. Funds flow through official Alipay channels.
type alipayProvider struct {
	client *alipay.Client
	cfg    AlipayConfig
	logger *zap.Logger
}

// NewAlipayProvider creates an Alipay provider.
func NewAlipayProvider(cfg AlipayConfig, logger *zap.Logger) (Provider, error) {
	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return nil, fmt.Errorf("init alipay client: %w", err)
	}

	client.SetNotifyUrl(cfg.NotifyURL)
	client.SetReturnUrl(cfg.ReturnURL)

	// Set Alipay public key for automatic response signature verification.
	if cfg.AlipayPublicKey != "" {
		client.AutoVerifySign([]byte(cfg.AlipayPublicKey))
	}

	return &alipayProvider{
		client: client,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (p *alipayProvider) Name() string { return "alipay" }

// CreateOrder creates a payment order using Alipay's precreate API (QR code).
func (p *alipayProvider) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.OrderID).
		Set("total_amount", req.Amount.StringFixed(2)).
		Set("subject", req.Description)

	rsp, err := p.client.TradePrecreate(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("alipay precreate: %w", err)
	}
	if rsp.Response.Code != "10000" {
		return nil, fmt.Errorf("alipay precreate failed: code=%s, msg=%s, sub_msg=%s",
			rsp.Response.Code, rsp.Response.Msg, rsp.Response.SubMsg)
	}

	return &CreateOrderResponse{
		OrderID:   req.OrderID,
		QRCodeURL: rsp.Response.QrCode,
		Channel:   ChannelAlipayQR,
	}, nil
}

// QueryOrder queries Alipay payment order status.
func (p *alipayProvider) QueryOrder(ctx context.Context, orderID string) (*QueryOrderResponse, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", orderID)

	rsp, err := p.client.TradeQuery(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("alipay query: %w", err)
	}
	if rsp.Response.Code != "10000" {
		return nil, fmt.Errorf("alipay query failed: code=%s, msg=%s", rsp.Response.Code, rsp.Response.Msg)
	}

	status := mapAlipayTradeStatus(rsp.Response.TradeStatus)
	amount, _ := decimal.NewFromString(rsp.Response.TotalAmount)

	return &QueryOrderResponse{
		OrderID:    rsp.Response.OutTradeNo,
		ExternalID: rsp.Response.TradeNo,
		Amount:     amount,
		Status:     status,
		Channel:    ChannelAlipayQR,
		PaidAt:     rsp.Response.SendPayDate,
	}, nil
}

// Transfer initiates an Alipay fund transfer (for creator withdrawal).
func (p *alipayProvider) Transfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_biz_no", req.OrderID).
		Set("trans_amount", req.Amount.StringFixed(2)).
		Set("biz_scene", "DIRECT_TRANSFER").
		Set("product_code", "TRANS_ACCOUNT_NO_PWD").
		Set("order_title", req.Description).
		SetBodyMap("payee_info", func(bm gopay.BodyMap) {
			bm.Set("identity", req.Account).
				Set("identity_type", "ALIPAY_LOGON_ID").
				Set("name", req.AccountName)
		})

	rsp, err := p.client.FundTransUniTransfer(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("alipay transfer: %w", err)
	}
	if rsp.Response.Code != "10000" {
		return nil, fmt.Errorf("alipay transfer failed: code=%s, msg=%s, sub_msg=%s",
			rsp.Response.Code, rsp.Response.Msg, rsp.Response.SubMsg)
	}

	return &TransferResponse{
		OrderID:    req.OrderID,
		ExternalID: rsp.Response.OrderId,
		Status:     "completed",
	}, nil
}

// ParseNotify parses an Alipay async payment notification.
func (p *alipayProvider) ParseNotify(_ context.Context, r *http.Request) (*NotifyResult, error) {
	notifyReq, err := alipay.ParseNotifyToBodyMap(r)
	if err != nil {
		return nil, fmt.Errorf("parse alipay notify: %w", err)
	}

	status := mapAlipayTradeStatus(notifyReq.Get("trade_status"))
	amount, _ := decimal.NewFromString(notifyReq.Get("total_amount"))

	return &NotifyResult{
		OrderID:    notifyReq.Get("out_trade_no"),
		ExternalID: notifyReq.Get("trade_no"),
		Amount:     amount,
		Status:     status,
		Channel:    ChannelAlipayQR,
	}, nil
}

// VerifyNotify verifies an Alipay async notification signature.
func (p *alipayProvider) VerifyNotify(_ context.Context, r *http.Request) (bool, error) {
	notifyReq, err := alipay.ParseNotifyToBodyMap(r)
	if err != nil {
		return false, fmt.Errorf("parse alipay notify for verify: %w", err)
	}

	ok, err := alipay.VerifySign(p.cfg.AlipayPublicKey, notifyReq)
	if err != nil {
		return false, nil
	}
	return ok, nil
}

// mapAlipayTradeStatus maps Alipay trade status to internal order status.
func mapAlipayTradeStatus(status string) string {
	switch status {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return OrderStatusPaid
	case "WAIT_BUYER_PAY":
		return OrderStatusPending
	case "TRADE_CLOSED":
		return OrderStatusClosed
	default:
		return OrderStatusPending
	}
}
