package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type RewardGormDAO struct {
	db *gorm.DB
}

var _ RewardDAO = (*RewardGormDAO)(nil)

func NewRewardGormDAO(db *gorm.DB) RewardDAO {
	return &RewardGormDAO{db: db}
}

func (r *RewardGormDAO) Insert(ctx context.Context, reward Reward) (int64, error) {
	now := time.Now().UnixMilli()
	reward.CreateTime = now
	reward.UpdateTime = now
	err := r.db.WithContext(ctx).Create(&reward).Error
	return reward.Id, err
}

func (r *RewardGormDAO) GetReward(ctx context.Context, rid int64) (Reward, error) {
	var reward Reward
	err := r.db.WithContext(ctx).Where("id = ?", rid).First(&reward).Error
	return reward, err
}

func (r *RewardGormDAO) UpdateStatus(ctx context.Context, rid int64, status uint8) error {
	return r.db.Model(&Reward{}).Where("id = ?", rid).Updates(map[string]any{
		"status":      status,
		"update_time": time.Now().UnixMilli(),
	}).Error
}
