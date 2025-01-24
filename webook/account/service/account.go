package service

import (
	"context"
	"webook/account/domain"
	"webook/account/repository"

	"github.com/to404hanga/pkg404/logger"
)

type accountService struct {
	repo repository.AccountRepository
	l    logger.Logger
}

var _ AccountService = (*accountService)(nil)

func NewAccountService(repo repository.AccountRepository, l logger.Logger) AccountService {
	return &accountService{
		repo: repo,
		l:    l,
	}
}

func (s *accountService) Credit(ctx context.Context, cr domain.Credit) error {
	err := s.repo.CheckUnique(ctx, cr)
	if err != nil {
		return err
	}
	err = s.repo.AddCredit(ctx, cr)
	if err == nil {
		er := s.repo.SetUnique(ctx, cr)
		if er != nil {
			s.l.Error("入账业务建立唯一索引失败", logger.Error(er), logger.String("biz", cr.Biz), logger.Int64("biz_id", cr.BizId))
		}
	}
	return err
}
