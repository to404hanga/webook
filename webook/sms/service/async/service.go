package async

import (
	"context"
	"time"
	"webook/sms/domain"
	"webook/sms/repository"
	"webook/sms/service"

	"github.com/to404hanga/pkg404/logger"
)

type Service struct {
	svc  service.Service
	repo repository.AsyncSmsRepository
	l    logger.Logger
}

var _ service.Service = (*Service)(nil)

func NewService(svc service.Service, repo repository.AsyncSmsRepository, l logger.Logger) *Service {
	res := &Service{
		svc:  svc,
		repo: repo,
		l:    l,
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// StartAsyncCycle 异步发送消息
func (s *Service) StartAsyncCycle() {
	// time.Sleep(time.Second * 3)
	for {
		s.AsyncSend()
	}
}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		// 执行发送
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			s.l.Error("执行异步发送短信失败", logger.Error(err), logger.Int64("id", as.Id))
		}
		res := err == nil
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败", logger.Error(err), logger.Bool("res", res), logger.Int64("id", as.Id))
		}
	case repository.ErrWaitingSMSNotFound:
		time.Sleep(time.Second)
	default:
		s.l.Error("抢占异步发送短信任务失败", logger.Error(err))
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	if s.needAsync() {
		err := s.repo.Add(ctx, domain.AsyncSms{
			TplId:    tplId,
			Args:     args,
			Numbers:  numbers,
			RetryMax: 3,
		})
		return err
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}

func (s *Service) needAsync() bool {
	// TODO
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
	return true
}
