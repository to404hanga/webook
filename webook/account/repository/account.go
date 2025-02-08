package repository

import (
	"context"
	"time"
	"webook/account/domain"
	"webook/account/repository/cache"
	"webook/account/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
)

type accountRepository struct {
	dao   dao.AccountDAO
	cache cache.AccountCache
}

var _ AccountRepository = (*accountRepository)(nil)

func NewAccountRepository(dao dao.AccountDAO, cache cache.AccountCache) AccountRepository {
	return &accountRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *accountRepository) CheckUnique(ctx context.Context, c domain.Credit) error {
	return repo.cache.GetUnique(ctx, c)
}

func (repo *accountRepository) SetUnique(ctx context.Context, c domain.Credit) error {
	return repo.cache.SetUnique(ctx, c)
}

func (repo *accountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	now := time.Now().UnixMilli()
	activities := transform.SliceFromSlice[domain.CreditItem, dao.AccountActivity](c.Items, func(idx int, ci domain.CreditItem) dao.AccountActivity {
		return dao.AccountActivity{
			Uid:         ci.Uid,
			Biz:         c.Biz,
			BizId:       c.BizId,
			Account:     ci.Account,
			AccountType: ci.AccountType.AsUint8(),
			Amount:      ci.Amt,
			Currency:    ci.Currency,
			CreateTime:  now,
			UpdateTime:  now,
		}
	})
	return repo.dao.AddActivities(ctx, activities...)
}
