package wechat

import (
	"context"
	"webook/payment/domain"
)

type PaymentService interface {
	// Prepay 预支付，对应微信创建订单的步骤
	Prepay(ctx context.Context, pmt domain.Payment) (string, error)
}
