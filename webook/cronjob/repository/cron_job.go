package repository

import (
	"context"
	"time"
	"webook/cronjob/domain"
	"webook/cronjob/repository/dao"
)

var ErrNoMoreJob = dao.ErrNoMoreJob

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	Release(ctx context.Context, jobId int64) error
	UpdateUpdateTime(ctx context.Context, jobId int64) error
	UpdateNextTime(ctx context.Context, jobId int64, nextTime time.Time) error
	AddJob(ctx context.Context, job domain.CronJob) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

var _ CronJobRepository = (*PreemptJobRepository)(nil)

func NewPreemptJobRepository() CronJobRepository {
	return &PreemptJobRepository{}
}

func (r *PreemptJobRepository) AddJob(ctx context.Context, job domain.CronJob) error {
	return r.dao.Insert(ctx, r.toEntity(job))
}

func (r *PreemptJobRepository) Preempt(ctx context.Context) (domain.CronJob, error) {
	job, err := r.dao.Preempt(ctx)
	return r.toDomain(job), err
}

func (r *PreemptJobRepository) Release(ctx context.Context, jobId int64) error {
	return r.dao.Release(ctx, jobId)
}

func (r *PreemptJobRepository) UpdateUpdateTime(ctx context.Context, jobId int64) error {
	return r.dao.UpdateUpdateTime(ctx, jobId)
}

func (r *PreemptJobRepository) UpdateNextTime(ctx context.Context, jobId int64, nextTime time.Time) error {
	return r.dao.UpdateNextTime(ctx, jobId, nextTime)
}

func (r *PreemptJobRepository) toEntity(job domain.CronJob) dao.Job {
	return dao.Job{
		Id:         job.Id,
		Name:       job.Name,
		Expression: job.Expression,
		Cfg:        job.Cfg,
		Executor:   job.Executor,
		NextTime:   job.NextTime.UnixMilli(),
	}
}

func (r *PreemptJobRepository) toDomain(job dao.Job) domain.CronJob {
	return domain.CronJob{
		Id:         job.Id,
		Name:       job.Name,
		Expression: job.Expression,
		Cfg:        job.Cfg,
		Executor:   job.Executor,
		NextTime:   time.UnixMilli(job.NextTime),
	}
}
