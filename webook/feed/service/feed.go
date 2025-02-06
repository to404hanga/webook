package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"

	"golang.org/x/sync/errgroup"
)

type feedService struct {
	repo         repository.FeedEventRepository
	handlerMap   map[string]Handler
	followClient followv1.FollowServiceClient
}

var _ FeedService = (*feedService)(nil)

func NewFeedService(repo repository.FeedEventRepository, handlerMap map[string]Handler, followClient followv1.FollowServiceClient) FeedService {
	return &feedService{
		repo:         repo,
		handlerMap:   handlerMap,
		followClient: followClient,
	}
}

func (f *feedService) RegisterService(_type string, handler Handler) {
	f.handlerMap[_type] = handler
}

func (f *feedService) CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error {
	handler, ok := f.handlerMap[feed.Type]
	if !ok {
		return fmt.Errorf("未能找到对应的 Handler %s", feed.Type)
	}
	return handler.CreateFeedEvent(ctx, feed.Ext)
}

func (f *feedService) GetFeedEventList(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	var (
		eg   errgroup.Group
		lock sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit*len(f.handlerMap))
	for _, handler := range f.handlerMap {
		h := handler
		eg.Go(func() error {
			evts, err := h.FindFeedEvents(ctx, uid, timestamp, limit)
			if err != nil {
				return err
			}
			lock.Lock()
			events = append(events, evts...)
			lock.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreateTime.UnixMilli() > events[j].CreateTime.UnixMilli()
	})
	return events[:min(limit, len(events))], nil
}
