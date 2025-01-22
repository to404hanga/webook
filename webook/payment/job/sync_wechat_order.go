package job

import (
	"context"
	"time"
	"webook/payment/service/wechat"

	"github.com/to404hanga/pkg404/logger"
)

type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.Logger
}

func NewSyncWechatOrderJob(svc *wechat.NativePaymentService, l logger.Logger) *SyncWechatOrderJob {
	return &SyncWechatOrderJob{
		svc: svc,
		l:   l,
	}
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

func (s *SyncWechatOrderJob) Run() error {
	t := time.Now().Add(-time.Minute * 16)
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, limit, offset, t)
		cancel()
		if err != nil {
			return err
		}
		for _, pmt := range pmts {
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
			err = s.svc.SyncWechatInfo(ctx, pmt.BizTradeNO)
			cancel()
			if err != nil {
				s.l.Error("同步微信订单状态失败", logger.Error(err), logger.String("biz_trade_no", pmt.BizTradeNO))
			}
		}
		if len(pmts) < limit {
			return nil
		}
		offset += len(pmts)
	}
}
