package repository

import (
	"context"
	"webook/search/domain"
	"webook/search/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
	"golang.org/x/sync/errgroup"
)

type articleRepository struct {
	articleDAO dao.ArticleDAO
	tagDAO     dao.TagDAO
	collectDAO dao.CollectDAO
	likeDAO    dao.LikeDAO
}

var _ ArticleRepository = (*articleRepository)(nil)

func NewArticleRepository(articleDAO dao.ArticleDAO, tagDAO dao.TagDAO, collectDAO dao.CollectDAO, likeDAO dao.LikeDAO) ArticleRepository {
	return &articleRepository{
		articleDAO: articleDAO,
		tagDAO:     tagDAO,
		collectDAO: collectDAO,
		likeDAO:    likeDAO,
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
	var (
		eg                                               errgroup.Group
		collectArticleIds, tagArticleIds, likeArticleIds []int64
		err                                              error
	)
	eg.Go(func() error {
		tagArticleIds, err = a.tagDAO.Search(ctx, uid, "article", keywords)
		return err
	})
	eg.Go(func() error {
		likeArticleIds, err = a.likeDAO.Search(ctx, uid, "article")
		return err
	})
	eg.Go(func() error {
		collectArticleIds, err = a.collectDAO.Search(ctx, uid, "article")
		return err
	})
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	articles, err := a.articleDAO.Search(ctx, dao.SearchReq{
		LikeIds:    likeArticleIds,
		TagIds:     tagArticleIds,
		CollectIds: collectArticleIds,
	}, keywords)
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[dao.Article, domain.Article](articles, func(idx int, src dao.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Status:  src.Status,
			Content: src.Content,
			Tags:    src.Tags,
		}
	}), nil
}
