package service

import (
	"context"
	"webook/feed/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/types.mock.go -package=svcmocks FeedService
type FeedService interface {
	CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error
	GetFeedEventList(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error)
}

// Handler 具体业务处理逻辑
//
//go:generate mockgen -source=./types.go -destination=./mocks/types.mock.go -package=svcmocks Handler
type Handler interface {
	CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error
	FindFeedEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error)
}
