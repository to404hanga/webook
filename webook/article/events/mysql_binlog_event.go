package events

import (
	"context"
	"time"
	"webook/article/domain"
	"webook/article/repository"
	"webook/article/repository/dao"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/canalx"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

type MySQLBinlogConsumer struct {
	client sarama.Client
	l      logger.Logger
	repo   *repository.CachedArticleRepository
}

func (m *MySQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("pub_articles_cache", m.client)
	if err != nil {
		return err
	}

	go func() {
		err = cg.Consume(context.Background(), []string{"webook_binlog"}, saramax.NewHandler[canalx.Message[dao.PublishedArticle]](m.l, m.Consume))
		if err != nil {
			m.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return err
}

func (m *MySQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[dao.PublishedArticle]) error {
	if val.Table != "published_articles" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, data := range val.Data {
		var err error
		switch data.Status {
		case domain.ArticleStatusPublished.ToUint8():
			err = m.repo.Cache().SetPub(ctx, m.repo.ToDomain(dao.Article(data)))
		case domain.ArticleStatusPrivate.ToUint8():
			err = m.repo.Cache().DelPub(ctx, data.Id)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
