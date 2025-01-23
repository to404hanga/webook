package dao

import "context"

//go:generate mockgen -source=./types.go -destination=./mocks/account.mock.go -package=daomocks AccountDAO
type AccountDAO interface {
	AddActivities(ctx context.Context, activities ...AccountActivity) error
}

type Account struct {
	Id         int64 `gorm:"primaryKey,autoIncrement"`
	Uid        int64
	Account    int64 `gorm:"uniqueIndex:account)type"`
	Type       uint8 `gorm:"uniqueIndex:account_type"`
	Balance    int64
	Currency   string
	UpdateTime int64
	CreateTime int64
}

type AccountActivity struct {
	Id          int64 `gorm:"primaryKey,autoIncrement"`
	Uid         int64
	Biz         string `gorm:"index:biz_type_id"`
	BizId       int64  `gorm:"index:biz_type_id"`
	Account     int64  `gorm:"index:account_type"`
	AccountType uint8  `gorm:"index:account_type"`
	Amount      int64
	Currency    string
	UpdateTime  int64
	CreateTime  int64
}

func (AccountActivity) TableName() string {
	return "account_activity"
}
