package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"webook/tag/domain"

	"github.com/redis/go-redis/v9"
	"github.com/to404hanga/pkg404/stl/transform"
)

var ErrKeyNotExist = redis.Nil

type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

var _ TagCache = (*RedisTagCache)(nil)

func NewRedisTagCache(client redis.Cmdable) TagCache {
	return &RedisTagCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) (res []domain.Tag, err error) {
	defer func() {
		if rcr := recover(); rcr != nil {
			res = nil
			err = fmt.Errorf("panic: %v", rcr)
		}
	}()

	data, err := r.client.HGetAll(ctx, r.userTagsKey(uid)).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrKeyNotExist
	}
	res = transform.SliceFromMap[string, string, domain.Tag](data, func(s1, s2 string) domain.Tag {
		var t domain.Tag
		err = json.Unmarshal([]byte(s2), &t)
		if err != nil {
			panic(err) // 直接 panic，交由 defer 处理
		}
		return t
	})
	return res, nil
}

func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	key := r.userTagsKey(uid)
	pipe := r.client.Pipeline()
	for _, tag := range tags {
		val, err := json.Marshal(tag)
		if err != nil {
			return err
		}
		pipe.HMSet(ctx, key, strconv.FormatInt(tag.Id, 10), val)
	}
	pipe.Expire(ctx, key, r.expiration)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisTagCache) DelTags(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.userTagsKey(uid)).Err()
}

func (r *RedisTagCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}
