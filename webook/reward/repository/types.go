package repository

import (
	"context"
	"webook/reward/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/reward.mock.go -package=repomocks RewardRepository
type RewardRepository interface {
	CreateReward(ctx context.Context, reward domain.Reward) (int64, error)
	GetReward(ctx context.Context, rid int64) (domain.Reward, error)
	GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CachedCodeURL(ctx context.Context, cu domain.CodeURL, r domain.Reward) error
	UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error
}
