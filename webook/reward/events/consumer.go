package events

import (
	"context"
	"strings"
	"time"
	"webook/reward/domain"
	"webook/reward/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

type PaymentEvent struct {
	BizTradeNO string
	Status     uint8
}

func (p PaymentEvent) ToDomainStatus() domain.RewardStatus {
	switch p.Status {
	case 1:
		return domain.RewardStatusInit
	case 2:
		return domain.RewardStatusPayed
	case 3:
		return domain.RewardStatusFailed
	default:
		return domain.RewardStatusUnknown
	}
}

type PaymentEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.RewardService
}

func (r *PaymentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("reward", r.client)
	if err != nil {
		return err
	}
	go func() {
		if err := cg.Consume(context.Background(), []string{"payment_events"}, saramax.NewHandler(r.l, r.Consume)); err != nil {
			r.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (r *PaymentEventConsumer) Consume(msg *sarama.ConsumerMessage, evt PaymentEvent) error {
	if !strings.HasPrefix(evt.BizTradeNO, "reward") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return r.svc.UpdateReward(ctx, evt.BizTradeNO, evt.ToDomainStatus())
}
