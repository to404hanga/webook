package failover

import (
	"context"
	"sync/atomic"
	"webook/internal/service/sms"
)

type TimeoutFailoverSMSService struct {
	svcs      []sms.Service
	idx       int64 // 当前使用的节点
	cnt       int64 // 连续超时数
	threshold int64 // 切换的阈值
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int64) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		idx:       0,
		cnt:       0,
		threshold: threshold,
	}
}

func (s *TimeoutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt64(&s.idx)
	cnt := atomic.LoadInt64(&s.cnt)
	if cnt > s.threshold {
		newIdx := (idx + 1) % int64(len(s.svcs))
		if atomic.CompareAndSwapInt64(&s.idx, idx, newIdx) {
			atomic.StoreInt64(&s.cnt, 0)
		}
		idx = newIdx
	}
	svc := s.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		atomic.StoreInt64(&s.cnt, 0)
		return nil
	case context.Canceled, context.DeadlineExceeded:
		atomic.AddInt64(&s.cnt, 1)
	default:
		// 如果强调一定是超时，那么就不增加
		// 如果是 EOF 之类的错误，你还可以考虑直接切换
	}
	return err
}
