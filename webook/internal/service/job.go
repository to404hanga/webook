package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"

	"github.com/to404hanga/pkg404/logger"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, job domain.Job) error
	// TODO 暴露 job 的 crud 方法
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.Logger
	refreshInterval time.Duration
}

var _ CronJobService = (*cronJobService)(nil)

func NewCronJobService(repo repository.CronJobRepository, l logger.Logger) CronJobService {
	return &cronJobService{
		repo: repo,
		l:    l,
	}
}

func (s *cronJobService) ResetNextTime(ctx context.Context, job domain.Job) error {
	nextTime := job.NextTime()
	return s.repo.UpdateNextTime(ctx, job.Id, nextTime)
}

func (s *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := s.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTicker(s.refreshInterval)
	go func() {
		for range ticker.C {
			s.refresh(job.Id)
		}
	}()
	job.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.repo.Release(ctx, job.Id); err != nil {
			s.l.Error("释放 job 失败", logger.Error(err), logger.Int64("job_id", job.Id))
		}
	}
	return job, err
}

func (s *cronJobService) refresh(jobId int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := s.repo.UpdateUpdateTime(ctx, jobId)
	if err != nil {
		s.l.Error("续约失败", logger.Error(err), logger.Int64("job_id", jobId))
	}
}
