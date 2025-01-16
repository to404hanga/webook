package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jobId int64) error
	UpdateUpdateTime(ctx context.Context, jobId int64) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

var _ CronJobRepository = (*PreemptJobRepository)(nil)

func NewPreemptJobRepository() CronJobRepository {
	return &PreemptJobRepository{}
}

func (r *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := r.dao.Preempt(ctx)
	return domain.Job{
		Id: job.Id,
	}, err
}

func (r *PreemptJobRepository) Release(ctx context.Context, jobId int64) error {
	return r.dao.Release(ctx, jobId)
}

func (r *PreemptJobRepository) UpdateUpdateTime(ctx context.Context, jobId int64) error {
	return r.dao.UpdateUpdateTime(ctx, jobId)
}
