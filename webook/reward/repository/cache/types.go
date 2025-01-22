package cache

import (
	"context"
	"webook/reward/domain"
)

//go:generate mockgen -source=./types.go -package=cachemocks -destination=./mocks/reward.mock.go RewardCache
type RewardCache interface {
	GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error
}
