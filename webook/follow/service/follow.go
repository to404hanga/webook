package service

import (
	"context"
	"webook/follow/domain"
	"webook/follow/repository"
)

type followRelationService struct {
	repo repository.FollowRepository
}

var _ FollowRelationService = (*followRelationService)(nil)

func NewFollowRelationService(repo repository.FollowRepository) FollowRelationService {
	return &followRelationService{
		repo: repo,
	}
}

func (svc *followRelationService) GetFollowee(ctx context.Context, follower int64, limit, offset int) ([]domain.FollowRelation, error) {
	return svc.repo.GetFollowee(ctx, follower, limit, offset)
}

func (svc *followRelationService) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	return svc.repo.FollowInfo(ctx, follower, followee)
}

func (svc *followRelationService) Follow(ctx context.Context, follower, followee int64) error {
	return svc.repo.AddFollowRelation(ctx, domain.FollowRelation{
		Follower: follower,
		Followee: followee,
	})
}

func (svc *followRelationService) CancelFollow(ctx context.Context, follower, followee int64) error {
	return svc.repo.InactiveFollowRelation(ctx, follower, followee)
}

func (svc *followRelationService) GetFollower(ctx context.Context, followee int64, limit, offset int) ([]domain.FollowRelation, error) {
	return svc.repo.GetFollower(ctx, followee, limit, offset)
}

func (svc *followRelationService) GetFollowStatic(ctx context.Context, followee int64) (domain.FollowStatics, error) {
	return svc.repo.GetFollowStatics(ctx, followee)
}
