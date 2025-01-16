package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jobId int64) error
	UpdateUpdateTime(ctx context.Context, jobId int64) error
	UpdateNextTime(ctx context.Context, jobId int64, nextTime time.Time) error
}

type GormJobDAO struct {
	db *gorm.DB
}

var _ JobDAO = (*GormJobDAO)(nil)

func NewGormJobDAO(db *gorm.DB) JobDAO {
	return &GormJobDAO{db: db}
}

// 乐观锁抢占事务
func (dao *GormJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var job Job
		now := time.Now().UnixMilli()
		err := db.Where("status = ? AND next_time < ?", jobStatusWaiting, now).First(&job).Error
		if err != nil {
			return job, err
		}
		res := db.Model(&Job{}).Where("id = ? AND version = ?", job.Id, job.Version).Updates(map[string]interface{}{
			"status":      jobStatusRunning,
			"version":     job.Version + 1,
			"update_time": now,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到，尝试抢下一个
			continue
		}
		return job, err
	}
}

func (dao *GormJobDAO) Release(ctx context.Context, jobId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jobId).Updates(map[string]interface{}{
		"status":      jobStatusWaiting,
		"update_time": now,
	}).Error
}

func (dao *GormJobDAO) UpdateUpdateTime(ctx context.Context, jobId int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jobId).Updates(map[string]interface{}{
		"update_time": now,
	}).Error
}

func (dao *GormJobDAO) UpdateNextTime(ctx context.Context, jobId int64, nextTime time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jobId).Updates(map[string]interface{}{
		"update_time": now,
		"next_time":   nextTime.UnixMilli(),
	}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string // cron 表达式
	Status     int    // 表达是不是可以抢占，有没有被人抢占
	Version    int
	NextTime   int64 `gorm:"index"`
	UpdateTime int64
	CreateTime int64
}

const (
	jobStatusWaiting = iota // 没有人抢占
	jobStatusRunning        // 已经被人抢占
	jobStatusPaused         // 不再需要被调度
)
