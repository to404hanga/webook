package repository

import (
	"context"
	"webook/article/domain"
	"webook/article/repository/dao"
)

//go:generate mockgen -source=./article_author.go -destination=./mocks/article_author.mock.go -package=repomocks ArticleAuthorRep
type ArticleAuthorRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}

type articleAuthorRepository struct {
	dao dao.ArticleDAO
}

var _ ArticleAuthorRepository = (*articleAuthorRepository)(nil)

func NewArticleAuthorRepository(dao dao.ArticleDAO) ArticleAuthorRepository {
	return &articleAuthorRepository{dao: dao}
}

func (a *articleAuthorRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	return a.dao.Insert(ctx, a.toEntity(article))
}

func (a *articleAuthorRepository) Update(ctx context.Context, article domain.Article) error {
	return a.dao.UpdateById(ctx, a.toEntity(article))
}

func (a *articleAuthorRepository) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	}
}
