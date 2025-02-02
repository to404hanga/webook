package dao

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

//go:generate mockgen -source=./async.go -package=daomocks -destination=./mocks/async.mock.go AsyncSmsDAO
type AsyncSmsDAO interface {
	Insert(ctx context.Context, s AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

const (
	asyncStatusWaiting uint8 = iota
	asyncStatusFailed
	asyncStatusSuccess
)

type GormAsyncSmsDAO struct {
	db *gorm.DB
}

var _ AsyncSmsDAO = (*GormAsyncSmsDAO)(nil)

func NewGormAsyncSmsDAO(db *gorm.DB) AsyncSmsDAO {
	return &GormAsyncSmsDAO{
		db: db,
	}
}

func (dao *GormAsyncSmsDAO) Insert(ctx context.Context, s AsyncSms) error {
	return dao.db.WithContext(ctx).Create(&s).Error
}

func (dao *GormAsyncSmsDAO) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	var s AsyncSms
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 为了避开一些偶发性的失败，只找 1 分钟前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{
			Strength: "UPDATE",
		}).Where("update_time < ? AND status = ?", endTime, asyncStatusWaiting).First(&s).Error
		if err != nil {
			return err
		}
		err = tx.Model(&AsyncSms{}).Where("id = ?", s.Id).Updates(map[string]interface{}{
			"retry_cnt":   gorm.Expr("retry_cnt + 1"),
			"update_time": now,
		}).Error
		return err
	})
	return s, err
}

func (dao *GormAsyncSmsDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).Where("id = ?", id).Updates(map[string]interface{}{
		"update_time": now,
		"status":      asyncStatusSuccess,
	}).Error
}

func (dao *GormAsyncSmsDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).Where("id = ? AND retry_cnt >= retry_max", id).Updates(map[string]interface{}{
		"update_time": now,
		"status":      asyncStatusFailed,
	}).Error
}

type AsyncSms struct {
	Id         int64
	Config     sqlx.JsonColumn[SmsConfig]
	RetryCnt   int
	RetryMax   int
	Status     uint8
	CreateTime int64
	UpdateTime int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
