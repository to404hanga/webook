package dao

import (
	"context"

	"gorm.io/gorm"
)

// 拉模型
//
//go:generate mockgen -source=./feed_pull_event.go -destination=./mocks/feed_pull_event.mock.go -package=daomocks FeedPullEventDAO
type FeedPullEventDAO interface {
	CreatePullEvent(ctx context.Context, event FeedPullEvent) error
	FindPullEventList(ctx context.Context, uids []int64, timestamp int64, limit int) ([]FeedPullEvent, error)
	FindPullEventListWithType(ctx context.Context, _type string, uids []int64, timestamp int64, limit int) ([]FeedPullEvent, error)
}

type FeedPullEvent struct {
	Id         int64 `gorm:"primaryKey,autoIncrement"`
	Uid        int64 `gorm:"index;column:uid"` // 发件人
	Type       string
	Content    string
	CreateTime int64
}

type feedPullEventDAO struct {
	db *gorm.DB
}

var _ FeedPullEventDAO = (*feedPullEventDAO)(nil)

func NewFeedPullEventDAO(db *gorm.DB) FeedPullEventDAO {
	return &feedPullEventDAO{db: db}
}

func (f *feedPullEventDAO) CreatePullEvent(ctx context.Context, event FeedPullEvent) error {
	return f.db.WithContext(ctx).Create(&event).Error
}

func (f *feedPullEventDAO) FindPullEventList(ctx context.Context, uids []int64, timestamp int64, limit int) ([]FeedPullEvent, error) {
	var res []FeedPullEvent
	err := f.db.WithContext(ctx).Model(&FeedPullEvent{}).Where("uid IN ? AND create_time < ?", uids, timestamp).Order("create_time desc").Limit(limit).Find(&res).Error
	return res, err
}

func (f *feedPullEventDAO) FindPullEventListWithType(ctx context.Context, _type string, uids []int64, timestamp int64, limit int) ([]FeedPullEvent, error) {
	var res []FeedPullEvent
	err := f.db.WithContext(ctx).Model(&FeedPullEvent{}).Where("uid IN ? AND create_time < ? AND type = ?", uids, timestamp, _type).Order("create_time desc").Limit(limit).Find(&res).Error
	return res, err
}
