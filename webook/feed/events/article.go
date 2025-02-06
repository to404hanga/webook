package events

import (
	"context"
	"strconv"
	"time"
	"webook/feed/domain"
	"webook/feed/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

const topicArticleEvent = "article_feed_event"

type ArticleFeedEvent struct {
	Uid int64
	Aid int64
}

type ArticleEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.FeedService
}

func NewArticleEventConsumer(client sarama.Client, l logger.Logger, svc service.FeedService) *ArticleEventConsumer {
	return &ArticleEventConsumer{
		svc:    svc,
		client: client,
		l:      l,
	}
}

func (a *ArticleEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("article_feed_event", a.client)
	if err != nil {
		return err
	}

	go func() {
		err = cg.Consume(context.Background(), []string{topicArticleEvent}, saramax.NewHandler[ArticleFeedEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (a *ArticleEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ArticleFeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: service.FollowEventName,
		Ext: map[string]string{
			"uid": strconv.FormatInt(evt.Uid, 10),
			"aid": strconv.FormatInt(evt.Aid, 10),
		},
	})
}
