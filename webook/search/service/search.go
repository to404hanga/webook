package service

import (
	"context"
	"strings"
	"webook/search/domain"
	"webook/search/repository"

	"golang.org/x/sync/errgroup"
)

type searchService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
}

var _ SearchService = (*searchService)(nil)

func NewSearchService(userRepo repository.UserRepository, articleRepo repository.ArticleRepository) SearchService {
	return &searchService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
	}
}

func (s *searchService) Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error) {
	keywords := strings.Split(expression, " ")

	var (
		eg  errgroup.Group
		res domain.SearchResult
	)

	eg.Go(func() error {
		users, err := s.userRepo.SearchUser(ctx, keywords)
		res.Users = users
		return err
	})
	eg.Go(func() error {
		articles, err := s.articleRepo.SearchArticle(ctx, uid, keywords)
		res.Articles = articles
		return err
	})

	return res, eg.Wait()
}
