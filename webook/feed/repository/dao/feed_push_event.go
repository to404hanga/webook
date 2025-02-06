package dao

import (
	"context"

	"gorm.io/gorm"
)

// 推模型
//
//go:generate mockgen -source=./feed_push_event.go -destination=./mocks/feed_push_event.mock.go -package=daomocks FeedPushEventDAO
type FeedPushEventDAO interface {
	CreatePushEvent(ctx context.Context, events []FeedPushEvent) error
	GetPushEvents(ctx context.Context, uid, timestamp int64, limit int) ([]FeedPushEvent, error)
	GetPushEventsWithType(ctx context.Context, _type string, uid, timestamp int64, limit int) ([]FeedPushEvent, error)
}

type FeedPushEvent struct {
	Id         int64 `gorm:"primaryKey,autoIncrement"`
	Uid        int64 `gorm:"index;column:uid"` // 收件人
	Type       string
	Content    string
	CreateTime int64
}

type feedPushEventDAO struct {
	db *gorm.DB
}

var _ FeedPushEventDAO = (*feedPushEventDAO)(nil)

func NewFeedPushEventDAO(db *gorm.DB) FeedPushEventDAO {
	return &feedPushEventDAO{db: db}
}

func (f *feedPushEventDAO) CreatePushEvent(ctx context.Context, events []FeedPushEvent) error {
	return f.db.WithContext(ctx).Create(&events).Error
}

func (f *feedPushEventDAO) GetPushEvents(ctx context.Context, uid, timestamp int64, limit int) ([]FeedPushEvent, error) {
	var res []FeedPushEvent
	err := f.db.WithContext(ctx).Model(&FeedPushEvent{}).Where("uid = ? AND create_time < ?", uid, timestamp).Order("create_time desc").Limit(limit).Find(&res).Error
	return res, err
}

func (f *feedPushEventDAO) GetPushEventsWithType(ctx context.Context, _type string, uid, timestamp int64, limit int) ([]FeedPushEvent, error) {
	var res []FeedPushEvent
	err := f.db.WithContext(ctx).Model(&FeedPushEvent{}).Where("uid = ? AND create_time < ? AND type = ?", uid, timestamp, _type).Order("create_time desc").Limit(limit).Find(&res).Error
	return res, err
}
