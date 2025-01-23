package service

import (
	"context"
	"webook/account/domain"
	"webook/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

var _ AccountService = (*accountService)(nil)

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	return s.repo.AddCredit(ctx, cr)
}
