package wechat

import (
	"context"
	"errors"
	"fmt"
	"time"
	"webook/payment/domain"
	"webook/payment/events"
	"webook/payment/repository"

	"github.com/to404hanga/pkg404/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
)

var ErrUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	appID     string
	mchID     string
	notifyURL string // 支付通知回调 URL
	repo      repository.PaymentRepository
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

var _ PaymentService = (*NativePaymentService)(nil)

func NewNativePaymentService(appID, mchID string, producer events.Producer, repo repository.PaymentRepository, svc *native.NativeApiService, l logger.Logger) *NativePaymentService {
	return &NativePaymentService{
		appID:     appID,
		mchID:     mchID,
		notifyURL: "http://wechat.meoying.com/pay/callback",
		repo:      repo,
		producer:  producer,
		svc:       svc,
		l:         l,
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
		},
	}
}

func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		OutTradeNo:  core.String(pmt.BizTradeNO),
		TimeExpire:  core.Time(time.Now().Add(time.Minute * 15)),
		Amount: &native.Amount{
			Total:    core.Int64(pmt.Amt.Total),
			Currency: core.String(pmt.Amt.Currency),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	// 对账
	txn, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNO),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, limit, offset int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, limit, offset, t)
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeId)
}

func (n *NativePaymentService) HandleCallBack(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, 微信的状态是 %s", ErrUnknownTransactionState, *txn.TradeState)
	}
	err := n.repo.UpdatePayment(ctx, domain.Payment{
		TxnID:      *txn.TransactionId,
		BizTradeNO: *txn.OutTradeNo,
		Status:     status,
	})
	if err != nil {
		return err
	}
	err = n.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
		BizTradeNO: *txn.OutTradeNo,
		Status:     status.AsUint8(),
	})
	if err != nil {
		n.l.Error("发送支付事件失败", logger.Error(err), logger.String("biz_trade_no", *txn.OutTradeNo))
	}
	return nil
}
