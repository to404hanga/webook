package service

import (
	"context"
	"time"
	"webook/feed/domain"
	"webook/feed/repository"
)

const LikeEventName = "like_event"

type LikeEventHandler struct {
	repo repository.FeedEventRepository
}

var _ Handler = (*LikeEventHandler)(nil)

func NewLikeEventHandler(repo repository.FeedEventRepository) Handler {
	return &LikeEventHandler{repo: repo}
}

func (l *LikeEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	return l.repo.FindPushEventsWithType(ctx, LikeEventName, uid, timestamp, limit)
}

func (l *LikeEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("liked").AsInt64()
	if err != nil {
		return err
	}
	return l.repo.CreatePushEvents(ctx, []domain.FeedEvent{{
		Uid:        uid,
		Ext:        ext,
		CreateTime: time.Now(),
		Type:       LikeEventName,
	}})
}
