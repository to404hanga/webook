package service

import (
	"context"
	"time"
	"webook/cronjob/domain"
	"webook/cronjob/repository"

	"github.com/to404hanga/pkg404/logger"
)

var ErrNoMoreJob = repository.ErrNoMoreJob

type CronJobService interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	ResetNextTime(ctx context.Context, job domain.CronJob) error
	AddJob(ctx context.Context, job domain.CronJob) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.Logger
	refreshInterval time.Duration
}

var _ CronJobService = (*cronJobService)(nil)

func NewCronJobService(repo repository.CronJobRepository, l logger.Logger) CronJobService {
	return &cronJobService{
		repo:            repo,
		l:               l,
		refreshInterval: time.Second * 10,
	}
}

func (s *cronJobService) AddJob(ctx context.Context, job domain.CronJob) error {
	job.NextTime = job.Next(time.Now())
	return s.repo.AddJob(ctx, job)
}

func (s *cronJobService) ResetNextTime(ctx context.Context, job domain.CronJob) error {
	nextTime := job.Next(time.Now())
	if !nextTime.IsZero() {
		return s.repo.UpdateNextTime(ctx, job.Id, nextTime)
	}
	return nil
}

func (s *cronJobService) Preempt(ctx context.Context) (domain.CronJob, error) {
	job, err := s.repo.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}

	ch := make(chan struct{})
	go func() {
		ticker := time.NewTicker(s.refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ch:
				// 退出续约循环
				return
			case <-ticker.C:
				s.refresh(job.Id)
			}
		}
	}()

	job.CancelFunc = func() {
		close(ch)
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
