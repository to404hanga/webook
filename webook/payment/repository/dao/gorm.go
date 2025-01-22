package dao

import (
	"context"
	"time"
	"webook/payment/domain"

	"gorm.io/gorm"
)

type PaymentGormDAO struct {
	db *gorm.DB
}

var _ PaymentDAO = (*PaymentGormDAO)(nil)

func NewPaymentGormDAO(db *gorm.DB) PaymentDAO {
	return &PaymentGormDAO{
		db: db,
	}
}

func (p *PaymentGormDAO) GetPayment(ctx context.Context, bizTradeNo string) (Payment, error) {
	var res Payment
	err := p.db.WithContext(ctx).Where("biz_trade_no = ?", bizTradeNo).First(&res).Error
	return res, err
}

func (p *PaymentGormDAO) FindExpiredPayment(ctx context.Context, limit, offset int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := p.db.WithContext(ctx).Where("status = ? AND update_time < ?", domain.PaymentStatusInit.AsUint8(), t.UnixMilli()).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (p *PaymentGormDAO) UpdateTxnIDAndStatus(ctx context.Context, bizTradeNo, txnID string, status domain.PaymentStatus) error {
	return p.db.WithContext(ctx).Model(&Payment{}).Where("biz_trade_no = ?", bizTradeNo).Updates(map[string]any{
		"txn_id":      txnID,
		"status":      status.AsUint8(),
		"update_time": time.Now().UnixMilli(),
	}).Error
}

func (p *PaymentGormDAO) Insert(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.UpdateTime = now
	pmt.CreateTime = now
	return p.db.WithContext(ctx).Create(&pmt).Error
}
