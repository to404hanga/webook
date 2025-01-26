package cache

import (
	"context"
	"fmt"
	"strconv"
	"webook/follow/domain"

	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

const (
	fieldFollowerCount = "follower_count"
	fieldFolloweeCount = "followee_count"
)

type FollowRedisCache struct {
	client redis.Cmdable
}

var _ FollowCache = (*FollowRedisCache)(nil)

func NewFollowRedisCache(client redis.Cmdable) FollowCache {
	return &FollowRedisCache{client: client}
}

func (f *FollowRedisCache) Follow(ctx context.Context, follower, followee int64) error {
	return f.updateStaticsInfo(ctx, follower, followee, 1)
}

func (f *FollowRedisCache) CancelFollow(ctx context.Context, follower, followee int64) error {
	return f.updateStaticsInfo(ctx, follower, followee, -1)
}

func (f *FollowRedisCache) StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	data, err := f.client.HGetAll(ctx, f.staticsKey(uid)).Result()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	if len(data) == 0 {
		return domain.FollowStatics{}, ErrKeyNotExist
	}
	var res domain.FollowStatics
	res.Followers, _ = strconv.ParseInt(data[fieldFollowerCount], 10, 64)
	res.Followees, _ = strconv.ParseInt(data[fieldFolloweeCount], 10, 64)
	return res, nil
}

func (f *FollowRedisCache) SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error {
	return f.client.HMSet(ctx, f.staticsKey(uid), fieldFollowerCount, statics.Followers, fieldFolloweeCount, statics.Followees).Err()
}

func (f *FollowRedisCache) updateStaticsInfo(ctx context.Context, follower, followee, delta int64) error {
	tx := f.client.TxPipeline()
	tx.HIncrBy(ctx, f.staticsKey(follower), fieldFolloweeCount, delta)
	tx.HIncrBy(ctx, f.staticsKey(followee), fieldFollowerCount, delta)
	_, err := tx.Exec(ctx)
	return err
}

func (f *FollowRedisCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}
