package service

import (
	"context"
	"webook/search/domain"
	"webook/search/repository"
)

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
	anyRepo     repository.AnyRepository
}

var _ SyncService = (*syncService)(nil)

func NewSyncService(userRepo repository.UserRepository, articleRepo repository.ArticleRepository, anyRepo repository.AnyRepository) SyncService {
	return &syncService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		anyRepo:     anyRepo,
	}
}

func (s *syncService) InputAny(ctx context.Context, indexName, docId, data string) error {
	return s.anyRepo.Input(ctx, indexName, docId, data)
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *syncService) Delete(ctx context.Context, indexName, docId string) error {
	return s.anyRepo.Delete(ctx, indexName, docId)
}
