package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"webook/feed/domain"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
)

var FolloweesNotFound = cache.FolloweesNotFound

//go:generate mockgen -source=./feed_event.go -destination=./mocks/feed_event.mock.go -package=repomocks FeedEventRepository
type FeedEventRepository interface {
	CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error
	CreatePullEvent(ctx context.Context, events domain.FeedEvent) error
	FindPullEvents(ctx context.Context, uids []int64, timestamp int64, limit int) ([]domain.FeedEvent, error)
	FindPushEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error)
	FindPullEventsWithType(ctx context.Context, _type string, uids []int64, timestamp int64, limit int) ([]domain.FeedEvent, error)
	FindPushEventsWithType(ctx context.Context, _type string, uid, timestamp int64, limit int) ([]domain.FeedEvent, error)
}

type CachedFeedEventRepository struct {
	pullDAO   dao.FeedPullEventDAO
	pushDAO   dao.FeedPushEventDAO
	feedCache cache.FeedEventCache
}

var _ FeedEventRepository = (*CachedFeedEventRepository)(nil)

func NewCachedFeedEventRepository(pullDAO dao.FeedPullEventDAO, pushDAO dao.FeedPushEventDAO, feedCache cache.FeedEventCache) FeedEventRepository {
	return &CachedFeedEventRepository{
		pullDAO:   pullDAO,
		pushDAO:   pushDAO,
		feedCache: feedCache,
	}
}

func (c *CachedFeedEventRepository) CreatePushEvents(ctx context.Context, events []domain.FeedEvent) error {
	pushEvents := make([]dao.FeedPushEvent, 0, len(events))
	for _, event := range events {
		pushEvents = append(pushEvents, c.toPushEventEntity(event))
	}
	return c.pushDAO.CreatePushEvent(ctx, pushEvents)
}

func (c *CachedFeedEventRepository) CreatePullEvent(ctx context.Context, event domain.FeedEvent) error {
	return c.pullDAO.CreatePullEvent(ctx, c.toPullEventEntity(event))
}

func (c *CachedFeedEventRepository) FindPullEvents(ctx context.Context, uids []int64, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	events, err := c.pullDAO.FindPullEventList(ctx, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := transform.SliceFromSlice[dao.FeedPullEvent, domain.FeedEvent](events, func(fpe dao.FeedPullEvent) domain.FeedEvent {
		return c.toPullEventDomain(fpe)
	})
	return ans, nil
}

func (c *CachedFeedEventRepository) FindPushEvents(ctx context.Context, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	events, err := c.pushDAO.GetPushEvents(ctx, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := transform.SliceFromSlice[dao.FeedPushEvent, domain.FeedEvent](events, func(fpe dao.FeedPushEvent) domain.FeedEvent {
		return c.toPushEventDomain(fpe)
	})
	return ans, nil
}

func (c *CachedFeedEventRepository) SetFollowees(ctx context.Context, follower int64, followees []int64) error {
	return c.feedCache.SetFollowees(ctx, follower, followees)
}

func (c *CachedFeedEventRepository) GetFollowees(ctx context.Context, follower int64) ([]int64, error) {
	followees, err := c.feedCache.GetFollowees(ctx, follower)
	if errors.Is(err, cache.FolloweesNotFound) {
		return nil, FolloweesNotFound
	}
	return followees, err
}

func (c *CachedFeedEventRepository) FindPullEventsWithType(ctx context.Context, _type string, uids []int64, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	events, err := c.pullDAO.FindPullEventListWithType(ctx, _type, uids, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := transform.SliceFromSlice[dao.FeedPullEvent, domain.FeedEvent](events, func(fpe dao.FeedPullEvent) domain.FeedEvent {
		return c.toPullEventDomain(fpe)
	})
	return ans, nil
}

func (c *CachedFeedEventRepository) FindPushEventsWithType(ctx context.Context, _type string, uid, timestamp int64, limit int) ([]domain.FeedEvent, error) {
	events, err := c.pushDAO.GetPushEventsWithType(ctx, _type, uid, timestamp, limit)
	if err != nil {
		return nil, err
	}
	ans := transform.SliceFromSlice[dao.FeedPushEvent, domain.FeedEvent](events, func(fpe dao.FeedPushEvent) domain.FeedEvent {
		return c.toPushEventDomain(fpe)
	})
	return ans, nil
}

func (c *CachedFeedEventRepository) toPushEventEntity(event domain.FeedEvent) dao.FeedPushEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPushEvent{
		Id:         event.Id,
		Uid:        event.Uid,
		Type:       event.Type,
		Content:    string(val),
		CreateTime: event.CreateTime.Unix(),
	}
}

func (c *CachedFeedEventRepository) toPullEventEntity(event domain.FeedEvent) dao.FeedPullEvent {
	val, _ := json.Marshal(event.Ext)
	return dao.FeedPullEvent{
		Id:         event.Id,
		Uid:        event.Uid,
		Type:       event.Type,
		Content:    string(val),
		CreateTime: event.CreateTime.Unix(),
	}
}

func (c *CachedFeedEventRepository) toPushEventDomain(event dao.FeedPushEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		Id:         event.Id,
		Uid:        event.Uid,
		Type:       event.Type,
		CreateTime: time.Unix(event.CreateTime, 0),
		Ext:        ext,
	}
}

func (c *CachedFeedEventRepository) toPullEventDomain(event dao.FeedPullEvent) domain.FeedEvent {
	var ext map[string]string
	_ = json.Unmarshal([]byte(event.Content), &ext)
	return domain.FeedEvent{
		Id:         event.Id,
		Uid:        event.Uid,
		Type:       event.Type,
		CreateTime: time.Unix(event.CreateTime, 0),
		Ext:        ext,
	}
}
