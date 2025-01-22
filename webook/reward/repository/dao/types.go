package dao

import "context"

//go:generate mockgen -source=./types.go -destination=./mocks/reward.mock.go -package=daomocks RewardDAO
type RewardDAO interface {
	Insert(ctx context.Context, r Reward) (int64, error)
	GetReward(ctx context.Context, rid int64) (Reward, error)
	UpdateStatus(ctx context.Context, rid int64, status uint8) error
}

type Reward struct {
	Id         int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Biz        string `gorm:"index:biz_biz_id"`
	BizId      int64  `gorm:"index:biz_biz_id"`
	BizName    string
	TargetUid  int64 `gorm:"index"`
	Status     uint8
	Uid        int64
	Amount     int64
	CreateTime int64
	UpdateTime int64
}
