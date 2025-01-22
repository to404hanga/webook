package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/reward/domain"

	"github.com/redis/go-redis/v9"
)

type RewardRedisCache struct {
	client redis.Cmdable
}

var _ RewardCache = (*RewardRedisCache)(nil)

func NewRewardRedisCache(client redis.Cmdable) RewardCache {
	return &RewardRedisCache{client: client}
}

func (r *RewardRedisCache) GetCachedCodeURL(ctx context.Context, reward domain.Reward) (domain.CodeURL, error) {
	key := r.codeURLKey(reward)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.CodeURL{}, err
	}
	var res domain.CodeURL
	err = json.Unmarshal(data, &res)
	return res, err
}

func (r *RewardRedisCache) CachedCodeURL(ctx context.Context, cu domain.CodeURL, reward domain.Reward) error {
	key := r.codeURLKey(reward)
	data, err := json.Marshal(cu)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, time.Minute*30).Err()
}

func (r *RewardRedisCache) codeURLKey(reward domain.Reward) string {
	return fmt.Sprintf("reward:code_url:%s:%d:%d", reward.Target.Biz, reward.Target.BizId, reward.Uid)
}
