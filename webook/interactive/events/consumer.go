package events

import (
	"context"
	"time"
	"webook/interactive/repository"

	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"

	"github.com/IBM/sarama"
)

const TopicReadEvent = "article_read"

type Consumer interface {
	Start() error
}

type ReadEvent struct {
	Aid int64
	Uid int64
}

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.Logger
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client, l logger.Logger) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client, l: l}
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewBatchHandler[ReadEvent](i.BatchConsume, i.l))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}

func (i *InteractiveReadEventConsumer) StartV1() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return err
}
func (i *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIds)
}

func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)
}
