package events

import (
	"context"
	"time"
	"webook/search/domain"
	"webook/search/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

const topicSyncUser = "sync_user_event"

type UserEvent struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}

type UserConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.Logger
}

var _ Consumer = (*UserConsumer)(nil)

func NewUserConsumer(syncSvc service.SyncService, client sarama.Client, l logger.Logger) *UserConsumer {
	return &UserConsumer{
		syncSvc: syncSvc,
		client:  client,
		l:       l,
	}
}

func (u *UserConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_user", u.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(), []string{topicSyncUser}, saramax.NewHandler(u.l, u.Consume))
		if err != nil {
			u.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (u *UserConsumer) Consume(sg *sarama.ConsumerMessage, evt UserEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return u.syncSvc.InputUser(ctx, u.toDomain(evt))
}

func (u *UserConsumer) toDomain(evt UserEvent) domain.User {
	return domain.User{
		Id:       evt.Id,
		Email:    evt.Email,
		Phone:    evt.Phone,
		Nickname: evt.Nickname,
	}
}
