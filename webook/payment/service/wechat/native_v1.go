package wechat

import (
	"context"
	"encoding/json"
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

type NativePaymentServiceV1 struct {
	baseNativePaymentService
	repo    *repository.PaymentGormRepository
	msgRepo *repository.LocalMsgGormRepository
}

var _ PaymentService = (*NativePaymentServiceV1)(nil)

func NewNativePaymentServiceV1(appID, mchID string, producer events.Producer, repo *repository.PaymentGormRepository, msgRepo *repository.LocalMsgGormRepository, svc *native.NativeApiService, l logger.Logger) *NativePaymentServiceV1 {
	return &NativePaymentServiceV1{
		baseNativePaymentService: baseNativePaymentService{
			appID:     appID,
			mchID:     mchID,
			notifyURL: "http://wechat.meoying.com/pay/callback",
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
		},
		repo:    repo,
		msgRepo: msgRepo,
	}
}

func (n *NativePaymentServiceV1) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
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

func (n *NativePaymentServiceV1) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
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

func (n *NativePaymentServiceV1) FindExpiredPayment(ctx context.Context, limit, offset int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, limit, offset, t)
}

func (n *NativePaymentServiceV1) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeId)
}

func (n *NativePaymentServiceV1) HandleCallBack(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

// updateByTxn 确保消息至少成功发送一次的版本
func (n *NativePaymentServiceV1) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, 微信的状态是 %s", errUnknownTransactionState, *txn.TradeState)
	}
	evt := events.PaymentEvent{
		BizTradeNO: *txn.OutTradeNo,
		Status:     status.AsUint8(),
	}
	var msgId int64
	err := n.repo.Transaction(ctx, func(pmt *repository.PaymentGormRepository, msg *repository.LocalMsgGormRepository) error {
		er := pmt.UpdatePayment(ctx, domain.Payment{
			TxnID:      *txn.TransactionId,
			BizTradeNO: *txn.OutTradeNo,
			Status:     status,
		})
		if er != nil {
			return er
		}
		evtData, er := json.Marshal(evt)
		if er != nil {
			return er
		}
		msgId, er = msg.AddMsg(ctx, string(evtData))
		return er
	})
	if err != nil {
		return err
	}

	err = n.producer.ProducePaymentEvent(ctx, evt)
	if err != nil {
		n.l.Error("发送支付事件失败", logger.Error(err), logger.String("biz_trade_no", *txn.OutTradeNo))
		return nil
	}

	// 更新本地消息表状态
	err = n.msgRepo.MarkSuccess(ctx, msgId)
	if err != nil {
		n.l.Error("标记本地消息表状态为成功失败", logger.Error(err), logger.Int64("msg_id", msgId), logger.String("biz_trade_no", *txn.OutTradeNo))
	}
	return nil
}
