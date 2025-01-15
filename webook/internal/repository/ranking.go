package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

//go:generate mockgen -source=./ranking.go -package=repomocks -destination=./mocks/ranking.mock.go RankingRepository
type RankingRepository interface {
	ReplaceTopN(ctx context.Context, articles []domain.Article) error
}

type CachedRankingRepository struct {
	cache cache.RankingCache
}

var _ RankingRepository = (*CachedRankingRepository)(nil)

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, articles []domain.Article) error {
	return c.cache.Set(ctx, articles)
}
