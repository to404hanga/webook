package repository

import (
	"context"
	"webook/search/domain"
	"webook/search/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
)

type articleRepository struct {
	articleDAO dao.ArticleDAO
	tagDAO     dao.TagDAO
}

var _ ArticleRepository = (*articleRepository)(nil)

func NewArticleRepository(articleDAO dao.ArticleDAO, tagDAO dao.TagDAO) ArticleRepository {
	return &articleRepository{
		articleDAO: articleDAO,
		tagDAO:     tagDAO,
	}
}

func (a *articleRepository) InputArticle(ctx context.Context, msg domain.Article) error {
	return a.articleDAO.InputArticle(ctx, dao.Article{
		Id:      msg.Id,
		Title:   msg.Title,
		Status:  msg.Status,
		Content: msg.Content,
	})
}

func (a *articleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error) {
	articleIds, err := a.tagDAO.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	articles, err := a.articleDAO.Search(ctx, articleIds, keywords)
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[dao.Article, domain.Article](articles, func(a dao.Article) domain.Article {
		return domain.Article{
			Id:      a.Id,
			Title:   a.Title,
			Status:  a.Status,
			Content: a.Content,
		}
	}), nil
}
