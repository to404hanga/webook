package events

import (
	"context"
	"time"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/canalx"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

type MySQLBinLogConsumer struct {
	client sarama.Client
	l      logger.Logger
	cache  cache.FollowCache
}

func (m *MySQLBinLogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("follow", m.client)
	if err != nil {
		return err
	}

	go func() {
		err = cg.Consume(context.Background(), []string{"webook_binlog"}, saramax.NewHandler[canalx.Message[FollowRelation]](m.l, m.Consume))
		if err != nil {
			m.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (m *MySQLBinLogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[FollowRelation]) error {
	if val.Table != "user" || val.Type != "INSERT" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	for _, data := range val.Data {
		var err error
		switch data.Status {
		case dao.FollowRelationStatusActive:
			err = m.cache.Follow(ctx, data.Follower, data.Followee)
		case dao.FollowRelationStatusInactive:
			err = m.cache.CancelFollow(ctx, data.Follower, data.Followee)
		default:
			m.l.Error("未知的数据状态", logger.Int("follow_relation_status", int(data.Status)))
		}
		if err != nil {
			m.l.Error("关注或取消关注错误", logger.Error(err), logger.Int64("id", data.Id), logger.Int64("follower", data.Follower), logger.Int64("followee", data.Followee))
		}
	}
	return nil
}

type FollowRelation struct {
	Id         int64 `gorm:"column:id;autoIncrement;primaryKey"`
	Follower   int64 `gorm:"uniqueIndex:follower_followee"`
	Followee   int64 `gorm:"uniqueIndex:follower_followee"`
	Status     uint8
	CreateTime int64
	UpdateTime int64
}
