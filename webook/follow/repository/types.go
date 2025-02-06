package repository

import (
	"context"
	"webook/follow/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/follow.mock.go -package=repomocks FollowRepository
type FollowRepository interface {
	GetFollowee(ctx context.Context, follower int64, limit, offset int) ([]domain.FollowRelation, error)
	GetFollower(ctx context.Context, followee int64, limit, offset int) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error)
	AddFollowRelation(ctx context.Context, c domain.FollowRelation) error
	InactiveFollowRelation(ctx context.Context, follower, followee int64) error
	GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error)
}
