package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

//go:generate mockgen -source=./ranking.go -package=repomocks -destination=./mocks/ranking.mock.go RankingRepository
type RankingRepository interface {
	ReplaceTopN(ctx context.Context, articles []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	cache cache.RankingCache
}

type CachedRankingRepository_LocalAndRedis struct {
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

var (
	_ RankingRepository = (*CachedRankingRepository)(nil)
	_ RankingRepository = (*CachedRankingRepository_LocalAndRedis)(nil)
)

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func NewCachedRankingRepository_LocalAndRedis(redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepository_LocalAndRedis{
		redisCache: redisCache,
		localCache: localCache,
	}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, articles []domain.Article) error {
	return c.cache.Set(ctx, articles)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}

func (c *CachedRankingRepository_LocalAndRedis) GetTopN(ctx context.Context) ([]domain.Article, error) {
	res, err := c.localCache.Get(ctx)
	if err == nil {
		return res, nil
	}

	res, err = c.redisCache.Get(ctx)
	if err != nil {
		return c.localCache.ForceGet(ctx)
	}

	_ = c.localCache.Set(ctx, res)

	return res, err
}

func (c *CachedRankingRepository_LocalAndRedis) ReplaceTopN(ctx context.Context, articles []domain.Article) error {
	_ = c.localCache.Set(ctx, articles)
	return c.redisCache.Set(ctx, articles)
}
