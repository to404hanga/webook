package service

import (
	"context"
	"time"
	"webook/feed/domain"
	"webook/feed/repository"
)

const FollowEventName = "follow_event"

type FollowEventHandler struct {
	repo repository.FeedEventRepository
}

var _ Handler = (*FollowEventHandler)(nil)

func NewFollowEventHandler(repo repository.FeedEventRepository) Handler {
	return &FollowEventHandler{
		repo: repo,
	}
}

func (f *FollowEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	return f.repo.FindPushEventsWithType(ctx, FollowEventName, uid, timestamp, limit)
}

func (f *FollowEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	followee, err := ext.Get("followee").AsInt64()
	if err != nil {
		return err
	}
	return f.repo.CreatePushEvents(ctx, []domain.FeedEvent{{
		Uid:        followee,
		Type:       FollowEventName,
		CreateTime: time.Now(),
		Ext:        ext,
	}})
}
