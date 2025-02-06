package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var FolloweesNotFound = redis.Nil

//go:generate mockgen -source=./feed_event.go -destination=./mocks/feed_event.mock.go -package=cachemocks FeedEventCache
type FeedEventCache interface {
	SetFollowees(ctx context.Context, follower int64, followees []int64) error
	GetFollowees(ctx context.Context, follower int64) ([]int64, error)
}

type FeedEventRedisCache struct {
	client redis.Cmdable
}

var _ FeedEventCache = (*FeedEventRedisCache)(nil)

func NewFeedEventRedisCache(client redis.Cmdable) FeedEventCache {
	return &FeedEventRedisCache{client: client}
}

const FolloweeKeyExpiration = 10 * time.Minute

func (f *FeedEventRedisCache) SetFollowees(ctx context.Context, follower int64, followees []int64) error {
	followeesBytes, err := json.Marshal(followees)
	if err != nil {
		return err
	}
	return f.client.Set(ctx, key(follower), string(followeesBytes), FolloweeKeyExpiration).Err()
}

func (f *FeedEventRedisCache) GetFollowees(ctx context.Context, follower int64) ([]int64, error) {
	res, err := f.client.Get(ctx, key(follower)).Result()
	if errors.Is(err, FolloweesNotFound) {
		return nil, FolloweesNotFound
	}
	if err != nil {
		return nil, err
	}
	var followees []int64
	err = json.Unmarshal([]byte(res), &followees)
	if err != nil {
		return nil, err
	}
	return followees, nil
}

func key(follower int64) string {
	return fmt.Sprintf("feed_event:%d", follower)
}
