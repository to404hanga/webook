package repository

import (
	"context"
	"time"
	"webook/account/domain"
	"webook/account/repository/dao"
)

type accountRepository struct {
	dao dao.AccountDAO
}

var _ AccountRepository = (*accountRepository)(nil)

func NewAccountRepository(dao dao.AccountDAO) AccountRepository {
	return &accountRepository{
		dao: dao,
	}
}

func (repo *accountRepository) AddCredit(ctx context.Context, c domain.Credit) error {
	activities := make([]dao.AccountActivity, 0, len(c.Items))
	now := time.Now().UnixMilli()
	for _, item := range c.Items {
		activities = append(activities, dao.AccountActivity{
			Uid:         item.Uid,
			Biz:         c.Biz,
			BizId:       c.BizId,
			Account:     item.Account,
			AccountType: item.AccountType.AsUint8(),
			Amount:      item.Amt,
			Currency:    item.Currency,
			CreateTime:  now,
			UpdateTime:  now,
		})
	}
	return repo.dao.AddActivities(ctx, activities...)
}
