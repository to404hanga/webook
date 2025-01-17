package job

import (
	"context"
	"sync"
	"time"
	"webook/internal/service"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/to404hanga/pkg404/logger"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger.Logger
	localLock sync.Mutex
	lock      *rlock.Lock
}

var _ Job = (*RankingJob)(nil)

func NewRankingJob(client *rlock.Client, svc service.RankingService, l logger.Logger, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		client:    client,
		key:       "job:ranking",
		l:         l,
		localLock: sync.Mutex{},
	}
}

// func (r *RankingJob) Run() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
// 	defer cancel()
// 	lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
// 		Interval: time.Millisecond * 100,
// 		Max:      3,
// 	}, time.Second)
// 	if err != nil {
// 		return err
// 	}
// 	defer func() {
// 		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 		defer cancel()
// 		err := lock.Unlock(ctx)
// 		if err != nil {
// 			r.l.Error("释放分布式锁失败", logger.Error(err))
// 		}
// 	}()

// 	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
// 	defer cancel()
// 	return r.svc.TopN(ctx)
// }

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	lock := r.lock

	if lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		var err error
		lock, err = r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		r.localLock.Unlock()
		go func() {
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

func (r *RankingJob) Name() string {
	return "ranking"
}
