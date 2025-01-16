package job

import (
	"context"
	"fmt"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/logger"

	"golang.org/x/sync/semaphore"
)

// Executor 任务执行器
type Executor interface {
	Name() string
	Exec(ctx context.Context, job domain.Job) error
}

// LocalFuncExecutor 调用本地方法的任务执行器
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, job domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, job domain.Job) error),
	}
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, job domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, job domain.Job) error {
	fn, ok := l.funcs[job.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", job.Name)
	}
	return fn(ctx, job)
}

type Scheduler struct {
	dbTimeout time.Duration
	svc       service.CronJobService
	executors map[string]Executor
	l         logger.Logger
	limiter   *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.Logger) *Scheduler {
	return &Scheduler{
		svc:       svc,
		l:         l,
		executors: make(map[string]Executor),
		limiter:   semaphore.NewWeighted(100),
		dbTimeout: time.Second,
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return
		}

		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		job, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			continue
		}

		exec, ok := s.executors[job.Executor]
		if !ok {
			s.l.Error("找不到执行器", logger.Int64("job_id", job.Id), logger.String("executor", job.Executor))
			continue
		}

		go func() {
			defer func() {
				s.limiter.Release(1)
				job.CancelFunc()
			}()
			err := exec.Exec(ctx, job)
			if err != nil {
				s.l.Error("执行任务失败", logger.Int64("job_id", job.Id), logger.Error(err))
				return
			}
			err = s.svc.ResetNextTime(ctx, job)
			if err != nil {
				s.l.Error("重置任务下次执行时间失败", logger.Int64("job_id", job.Id), logger.Error(err))
			}
		}()
	}
}
