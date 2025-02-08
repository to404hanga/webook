package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"

	"github.com/to404hanga/pkg404/stl/transform"
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
		// TODO 可以设计一个兜底机制，直接丢到 push_event 里面
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

// GetFeedEventListV1 支持活跃用户的 GetFeedEventList 版本
func (f *feedService) GetFeedEventListV1(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	var (
		eg   errgroup.Group
		lock sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit<<1)

	eg.Go(func() error {
		// 活跃用户
		if f.isActiveUser(uid) {
			return nil
		}
		resp, err := f.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			Follower: uid,
			Limit:    10000,
		})
		if err != nil {
			return err
		}
		followeeIds := transform.SliceFromSlice[*followv1.FollowRelation, int64](resp.GetFollowRelations(), func(i int, fr *followv1.FollowRelation) int64 {
			return fr.GetFollowee()
		})
		evts, err := f.repo.FindPullEvents(ctx, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})
	eg.Go(func() error {
		evts, err := f.repo.FindPushEvents(ctx, uid, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreateTime.UnixMilli() > events[j].CreateTime.UnixMilli()
	})
	return events[:min(limit, len(events))], nil
}

func (f *feedService) isActiveUser(uid int64) bool {
	// 在实践中，是否是活跃用户，一般是离线任务计算的
	// 比如每天计算一批或间隔一段时间计算一批
	// 可以考虑采用连续登陆之类的方案
	// 在判定是否为活跃用户时，可以利用 redis 实现 bit array 或布隆过滤器
	return false
}
