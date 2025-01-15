package cache

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"

	"github.com/ecodeclub/ekit/syncx/atomicx"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

var _ RankingCache = (*RankingLocalCache)(nil)

func NewRankingLocalCache(expiration time.Duration) *RankingLocalCache {
	return &RankingLocalCache{
		expiration: expiration,
	}
}

func (r *RankingLocalCache) Set(ctx context.Context, articles []domain.Article) error {
	r.topN.Store(articles)
	r.ddl.Store(time.Now().Add(r.expiration))
	return nil
}

func (r *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	ddl := r.ddl.Load()
	articles := r.topN.Load()
	if len(articles) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存已失效")
	}
	return articles, nil
}

func (r *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	articles := r.topN.Load()
	if len(articles) == 0 {
		return nil, errors.New("本地缓存已失效")
	}
	return articles, nil
}
