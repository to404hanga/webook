package dao

import (
	"context"
	"database/sql"
	"time"
	"webook/payment/domain"
)

//go:generate mockgen -source=./types.go -package=daomocks -destination=./mocks/types.mock.go PaymentDAO
type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo, txnID string, status domain.PaymentStatus) error
	FindExpiredPayment(ctx context.Context, limit, offset int, t time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, bizTradeNo string) (Payment, error)
}

type Payment struct {
	Id          int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Amt         int64
	Currency    string
	Description string         `gorm:"description"`
	BizTradeNO  string         `gorm:"column:biz_trade_no;type:varchar(255);unique"`
	TxnID       sql.NullString `gorm:"txn_id;type:varchar(128);unique"`
	Status      uint8
	UpdateTime  int64
	CreateTime  int64
}
