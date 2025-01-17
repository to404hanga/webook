package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"

	"github.com/to404hanga/pkg404/logger"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	ListPub(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error)
	GetPubById(ctx context.Context, id, userId int64) (domain.Article, error)
	Save(ctx context.Context, article domain.Article) (int64, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]domain.Article, error)
	Withdraw(ctx context.Context, userId, id int64) error
	Publish(ctx context.Context, article domain.Article) (int64, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer
	l        logger.Logger
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer, l logger.Logger) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        l,
	}
}

var _ ArticleService = (*articleService)(nil)

func (svc *articleService) ListPub(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error) {
	return svc.repo.ListPub(ctx, start, limit, offset)
}

func (svc *articleService) GetPubById(ctx context.Context, id, userId int64) (domain.Article, error) {
	res, err := svc.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			er := svc.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: userId,
			})
			if er != nil {
				svc.l.Error("ArticleSvc GetPubById 发送 ReadEvent 失败", logger.Int64("aid", id), logger.Int64("uid", userId), logger.Error(er))
			}
		}
	}()
	return res, err
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}

func (svc *articleService) GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]domain.Article, error) {
	return svc.repo.GetByAuthor(ctx, userId, limit, offset)
}

func (svc *articleService) Withdraw(ctx context.Context, userId, id int64) error {
	return svc.repo.SyncStatus(ctx, userId, id, domain.ArticleStatusPrivate)
}

func (svc *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, article)
}

func (svc *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	if article.Id > 0 {
		err := svc.repo.Update(ctx, article)
		return article.Id, err
	}
	return svc.repo.Create(ctx, article)
}
