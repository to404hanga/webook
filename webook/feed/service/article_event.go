package service

import (
	"context"
	"sort"
	"sync"
	"time"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/feed/domain"
	"webook/feed/repository"

	"github.com/to404hanga/pkg404/stl/transform"
	"golang.org/x/sync/errgroup"
)

const (
	ArticleEventName = "article_event"
	threshold        = 4
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepository
	followClient followv1.FollowServiceClient
}

var _ Handler = (*ArticleEventHandler)(nil)

func NewArticleEventHandler(repo repository.FeedEventRepository, followClient followv1.FollowServiceClient) Handler {
	return &ArticleEventHandler{
		repo:         repo,
		followClient: followClient,
	}
}

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("followee").AsInt64()
	if err != nil {
		return err
	}

	// 寻找被关注者的粉丝数，以判定使用拉模型或推模型
	resp, err := a.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{
		Followee: uid,
	})
	if err != nil {
		return err
	}

	now := time.Now()
	// 粉丝多，使用拉模型
	if resp.FollowStatic.Followers > threshold {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Uid:        uid,
			Type:       ArticleEventName,
			CreateTime: now,
			Ext:        ext,
		})
	} else {
		resp, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{
			Followee: uid,
		})
		if err != nil {
			return err
		}
		events := transform.SliceFromSlice[*followv1.FollowRelation, domain.FeedEvent](resp.GetFollowRelations(), func(fr *followv1.FollowRelation) domain.FeedEvent {
			return domain.FeedEvent{
				Uid:        fr.Follower,
				CreateTime: now,
				Type:       ArticleEventName,
				Ext:        ext,
			}
		})
		return a.repo.CreatePushEvents(ctx, events)
	}
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	var (
		eg   errgroup.Group
		lock sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit<<1)
	eg.Go(func() error {
		// 查询发件箱
		resp, err := a.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			Follower: uid,
			Limit:    10000,
		})
		if err != nil {
			return err
		}
		followeeIds := transform.SliceFromSlice[*followv1.FollowRelation, int64](resp.GetFollowRelations(), func(fr *followv1.FollowRelation) int64 {
			return fr.GetFollowee()
		})
		evts, err := a.repo.FindPullEventsWithType(ctx, ArticleEventName, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})
	eg.Go(func() error {
		evts, err := a.repo.FindPushEventsWithType(ctx, ArticleEventName, uid, timestamp, limit)
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
