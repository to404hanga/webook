package dao

import "context"

type FollowRelation struct {
	Id         int64 `gorm:"column:id;autoIncrement;primaryKey"`
	Follower   int64 `gorm:"uniqueIndex:follower_followee"`
	Followee   int64 `gorm:"uniqueIndex:follower_followee"`
	Status     uint8
	CreateTime int64
	UpdateTime int64
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

//go:generate mockgen -source=./types.go -destination=./mocks/follow.mock.go -package=daomocks FollowRelationDAO
type FollowRelationDAO interface {
	FollowRelationList(ctx context.Context, follower int64, limit, offset int) ([]FollowRelation, error)
	FansList(ctx context.Context, followee int64, limit, offset int) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower, followee int64) (FollowRelation, error)
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	UpdateStatus(ctx context.Context, followee, follower int64, status uint8) error
	CountFollower(ctx context.Context, uid int64) (int64, error)
	CountFollowee(ctx context.Context, uid int64) (int64, error)
}

type FollowStatics struct {
	Id         int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid        int64 `gorm:"unique"`
	Followers  int64
	Followees  int64
	UpdateTime int64
	CreateTime int64
}
