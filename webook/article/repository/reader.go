package repository

import (
	"context"
	"webook/article/domain"
	"webook/article/repository/dao"
)

//go:generate mockgen -source=./reader.go -destination=./mocks/reader.mock.go -package=repomocks ArticleReaderRepository
type ArticleReaderRepository interface {
	Save(ctx context.Context, article domain.Article) error
}

type articleReaderRepository struct {
	dao dao.ArticleReaderDAO
}

var _ ArticleReaderRepository = (*articleReaderRepository)(nil)

func NewArticleReaderRepository(dao dao.ArticleReaderDAO) ArticleReaderRepository {
	return &articleReaderRepository{dao: dao}
}

func (a *articleReaderRepository) Save(ctx context.Context, article domain.Article) error {
	return a.dao.Upsert(ctx, a.toEntity(article))
}

func (a *articleReaderRepository) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	}
}
