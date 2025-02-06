package repository

import (
	"context"
	"webook/follow/domain"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"

	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/stl/transform"
)

type CachedFollowRepository struct {
	dao   dao.FollowRelationDAO
	cache cache.FollowCache
	l     logger.Logger
}

var _ FollowRepository = (*CachedFollowRepository)(nil)

func NewCachedFollowRepository(dao dao.FollowRelationDAO, cache cache.FollowCache, l logger.Logger) FollowRepository {
	return &CachedFollowRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (repo *CachedFollowRepository) GetFollowee(ctx context.Context, follower int64, limit, offset int) ([]domain.FollowRelation, error) {
	followerList, err := repo.dao.FollowRelationList(ctx, follower, limit, offset)
	if err != nil {
		return nil, err
	}
	return repo.genFollowRelationList(followerList), nil
}

func (repo *CachedFollowRepository) GetFollower(ctx context.Context, followee int64, limit, offset int) ([]domain.FollowRelation, error) {
	followerList, err := repo.dao.FansList(ctx, followee, limit, offset)
	if err != nil {
		return nil, err
	}
	return repo.genFollowRelationList(followerList), nil
}

func (repo *CachedFollowRepository) FollowInfo(ctx context.Context, follower, followee int64) (domain.FollowRelation, error) {
	c, err := repo.dao.FollowRelationDetail(ctx, follower, followee)
	if err != nil {
		return domain.FollowRelation{}, err
	}
	return repo.toDomain(c), nil
}

func (repo *CachedFollowRepository) AddFollowRelation(ctx context.Context, c domain.FollowRelation) error {
	err := repo.dao.CreateFollowRelation(ctx, repo.toEntity(c))
	if err != nil {
		return err
	}
	return repo.cache.Follow(ctx, c.Follower, c.Followee)
}

func (repo *CachedFollowRepository) InactiveFollowRelation(ctx context.Context, follower, followee int64) error {
	err := repo.dao.UpdateStatus(ctx, followee, follower, dao.FollowRelationStatusInactive)
	if err != nil {
		return err
	}
	return repo.cache.CancelFollow(ctx, follower, followee)
}

func (repo *CachedFollowRepository) GetFollowStatics(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	res, err := repo.cache.StaticsInfo(ctx, uid)
	if err == nil {
		return res, nil
	}

	res.Followers, err = repo.dao.CountFollower(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}
	res.Followees, err = repo.dao.CountFollower(ctx, uid)
	if err != nil {
		return domain.FollowStatics{}, err
	}

	go func() {
		err = repo.cache.SetStaticsInfo(ctx, uid, res)
		if err != nil {
			repo.l.Warn("FollowRepo GetFollowStatics 设置缓存失败", logger.Error(err), logger.Int64("uid", uid))
		}
	}()

	return res, nil
}

func (repo *CachedFollowRepository) toDomain(fs dao.FollowRelation) domain.FollowRelation {
	return domain.FollowRelation{
		Follower: fs.Follower,
		Followee: fs.Followee,
	}
}

func (repo *CachedFollowRepository) toEntity(fs domain.FollowRelation) dao.FollowRelation {
	return dao.FollowRelation{
		Follower: fs.Follower,
		Followee: fs.Followee,
	}
}

func (repo *CachedFollowRepository) genFollowRelationList(followerList []dao.FollowRelation) []domain.FollowRelation {
	return transform.SliceFromSlice[dao.FollowRelation, domain.FollowRelation](followerList, func(fr dao.FollowRelation) domain.FollowRelation {
		return repo.toDomain(fr)
	})
}
