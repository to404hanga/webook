package cache

import (
	"context"
	"encoding/json"
	"time"
	"webook/internal/domain"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=./ranking.go -package=cachemocks -destination=./mocks/ranking.mock.go RankingCache
type RankingCache interface {
	Set(ctx context.Context, articles []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

var _ RankingCache = (*RankingRedisCache)(nil)

func NewRankingRedisCache(client redis.Cmdable) RankingCache {
	return &RankingRedisCache{
		client:     client,
		key:        "rankingL:top_n",
		expiration: time.Minute * 3,
	}
}

func (rc *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := rc.client.Get(ctx, rc.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (rc *RankingRedisCache) Set(ctx context.Context, articles []domain.Article) error {
	for i := range articles {
		articles[i].Content = articles[i].Abstract()
	}
	val, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, rc.key, val, rc.expiration).Err()
}
