package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webook/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailoverSMSService(svcs []sms.Service) *FailoverSMSService {
	return &FailoverSMSService{
		svcs: svcs,
		idx:  0,
	}
}

func (s *FailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// for _, svc := range s.svcs {
	// 	err := svc.Send(ctx, tplId, args, numbers...)
	// 	if err == nil {
	// 		return nil
	// 	}
	// 	log.Println(err)
	// }
	idx := atomic.AddUint64(&s.idx, 1)
	length := uint64(len(s.svcs))
	for i := idx; i < idx+length; i++ {
		svc := s.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询了所有服务商，但是都发送失败")
}
