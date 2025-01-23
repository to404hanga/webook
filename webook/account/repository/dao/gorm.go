package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountGormDAO struct {
	db *gorm.DB
}

var _ AccountDAO = (*AccountGormDAO)(nil)

func NewCreditGormDAO(db *gorm.DB) AccountDAO {
	return &AccountGormDAO{db: db}
}

func (dao *AccountGormDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		for _, activity := range activities {
			err := tx.Clauses(clause.OnConflict{
				DoUpdates: clause.Assignments(map[string]any{
					"balance":     gorm.Expr("`balance` + ?", activity.Amount),
					"update_time": now,
				}),
			}).Create(&Account{
				Uid:        activity.Uid,
				Account:    activity.Account,
				Type:       activity.AccountType,
				Balance:    activity.Amount,
				Currency:   activity.Currency,
				CreateTime: now,
				UpdateTime: now,
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Create(&activities).Error
	})
}
