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

const topicSyncArticle = "sync_article_event"

type ArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  int32  `json:"status"`
	Content string `json:"content"`
}

type ArticleConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.Logger
}

var _ Consumer = (*ArticleConsumer)(nil)

func NewArticleConsumer(syncSvc service.SyncService, client sarama.Client, l logger.Logger) *ArticleConsumer {
	return &ArticleConsumer{
		syncSvc: syncSvc,
		client:  client,
		l:       l,
	}
}

func (a *ArticleConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_article", a.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(), []string{topicSyncArticle}, saramax.NewHandler(a.l, a.Consume))
		if err != nil {
			a.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (a *ArticleConsumer) Consume(sg *sarama.ConsumerMessage, evt ArticleEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.syncSvc.InputArticle(ctx, a.toDomain(evt))
}

func (a *ArticleConsumer) toDomain(article ArticleEvent) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Status:  article.Status,
		Content: article.Content,
	}
}
