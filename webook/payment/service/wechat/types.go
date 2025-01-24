package wechat

import (
	"context"
	"webook/payment/domain"
	"webook/payment/events"

	"github.com/to404hanga/pkg404/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

//go:generate mockgen -source=./types.go -destination=./mocks/native.mock.go -package=svcmocks PaymentService
type PaymentService interface {
	// Prepay 预支付，对应微信创建订单的步骤
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}

type baseNativePaymentService struct {
	appID     string
	mchID     string
	notifyURL string // 支付通知回调 URL
	producer  events.Producer
	svc       *native.NativeApiService
	l         logger.Logger
	// 在微信 native 里面，分别是
	//
	// SUCCESS: 支付成功
	//
	// REFUND: 转入退款
	//
	// NOTPAY: 未支付
	//
	// CLOSED: 已关闭
	//
	// REVOKED: 已撤销（付款码支付）
	//
	// USERPAYING: 用户支付中（付款码支付）
	//
	// PAYERROR: 支付失败（其他原因，如银行返回失败）
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}
