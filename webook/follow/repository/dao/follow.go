package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FollowRelationGormDAO struct {
	db *gorm.DB
}

var _ FollowRelationDAO = (*FollowRelationGormDAO)(nil)

func NewFollowRelationGormDAO(db *gorm.DB) FollowRelationDAO {
	return &FollowRelationGormDAO{db: db}
}

func (f *FollowRelationGormDAO) CountFollower(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := f.db.WithContext(ctx).Model(FollowRelation{}).Select("count(follower)").Where("followee = ? AND status = ?", uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}

func (f *FollowRelationGormDAO) CountFollowee(ctx context.Context, uid int64) (int64, error) {
	var res int64
	err := f.db.WithContext(ctx).Model(FollowRelation{}).Select("count(followee)").Where("follower = ? AND status = ?", uid, FollowRelationStatusActive).Count(&res).Error
	return res, err
}

func (f *FollowRelationGormDAO) UpdateStatus(ctx context.Context, followee, follower int64, status uint8) error {
	return f.db.WithContext(ctx).Model(FollowRelation{}).Where("follower = ? AND followee = ?", follower, followee).Updates(map[string]any{
		"status":      status,
		"update_time": time.Now().UnixMilli(),
	}).Error
}

func (f *FollowRelationGormDAO) FollowRelationList(ctx context.Context, follower int64, limit, offset int) ([]FollowRelation, error) {
	var res []FollowRelation
	err := f.db.WithContext(ctx).Model(FollowRelation{}).Where("follower = ? AND status = ?", follower, FollowRelationStatusActive).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (f *FollowRelationGormDAO) FollowRelationDetail(ctx context.Context, follower, followee int64) (FollowRelation, error) {
	var res FollowRelation
	err := f.db.WithContext(ctx).Model(FollowRelation{}).Where("follower = ? AND followee = ? AND status = ?", follower, followee, FollowRelationStatusActive).First(&res).Error
	return res, err
}

func (f *FollowRelationGormDAO) CreateFollowRelation(ctx context.Context, c FollowRelation) error {
	now := time.Now().UnixMilli()
	c.CreateTime = now
	c.UpdateTime = now
	c.Status = FollowRelationStatusActive
	return f.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"status":      FollowRelationStatusActive,
			"update_time": now,
		}),
	}).Create(&c).Error
	// TODO 在这里更新 FollowStatics 的计数，也是 upsert 语义
}
