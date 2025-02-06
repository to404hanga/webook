package service

import (
	"context"
	"webook/follow/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/follow.mock.go -package=svcmocks FollowRelationService
type FollowRelationService interface {
	GetFollowee(ctx context.Context, follower int64, limit, offset int) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error)
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
	GetFollower(ctx context.Context, followee int64, limit, offset int) ([]domain.FollowRelation, error)
	GetFollowStatic(ctx context.Context, followee int64) (domain.FollowStatics, error)
}
