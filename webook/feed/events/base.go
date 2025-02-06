package events

import (
	"context"
	"time"
	"webook/feed/domain"
	"webook/feed/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

const topicFeedEvent = "feed_event"

type Consumer interface {
	Start() error
}

type FeedEvent struct {
	Type     string
	Metadata map[string]string // 使用 map[string]string，以解决使用 map[string]any 时传入 int64 反解析为 float64 的情况
}

type FeedEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.FeedService
}

func NewFeedEventConsumer(client sarama.Client, l logger.Logger, svc service.FeedService) *FeedEventConsumer {
	return &FeedEventConsumer{
		client: client,
		l:      l,
		svc:    svc,
	}
}

func (f *FeedEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("feed_event", f.client)
	if err != nil {
		return err
	}

	go func() {
		err = cg.Consume(context.Background(), []string{topicFeedEvent}, saramax.NewHandler[FeedEvent](f.l, f.Consume))
		if err != nil {
			f.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (f *FeedEventConsumer) Consume(msg *sarama.ConsumerMessage, evt FeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return f.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: evt.Type,
		Ext:  evt.Metadata,
	})
}
