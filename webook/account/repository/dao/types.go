package dao

import "context"

//go:generate mockgen -source=./types.go -destination=./mocks/account.mock.go -package=daomocks AccountDAO
type AccountDAO interface {
	AddActivities(ctx context.Context, activities ...AccountActivity) error
}

type Account struct {
	Id         int64 `gorm:"primaryKey,autoIncrement"`
	Uid        int64
	Account    int64 `gorm:"uniqueIndex:account_type"`
	Type       uint8 `gorm:"uniqueIndex:account_type"`
	Balance    int64
	Currency   string
	UpdateTime int64
	CreateTime int64
}

// 在 biz, biz_id, account, account_id 上创建一个联合唯一索引，可确保记账时不会重复记账
type AccountActivity struct {
	Id          int64 `gorm:"primaryKey,autoIncrement"`
	Uid         int64
	Biz         string `gorm:"uniqueIndex:biz_type_id"`
	BizId       int64  `gorm:"uniqueIndex:biz_type_id"`
	Account     int64  `gorm:"index:account_type;uniqueIndex:biz_type_id"`
	AccountType uint8  `gorm:"index:account_type;uniqueIndex:biz_type_id"`
	Amount      int64
	Currency    string
	UpdateTime  int64
	CreateTime  int64
}

func (AccountActivity) TableName() string {
	return "account_activity"
}
