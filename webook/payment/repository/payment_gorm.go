package repository

import (
	"context"
	"time"
	"webook/payment/domain"
	"webook/payment/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
	"gorm.io/gorm"
)

type PaymentGormRepository struct {
	dao dao.PaymentDAO
	db  *gorm.DB
}

var _ PaymentRepository = (*PaymentGormRepository)(nil)

func NewPaymentGormRepository(db *gorm.DB) *PaymentGormRepository {
	return &PaymentGormRepository{
		dao: dao.NewPaymentGormDAO(db),
		db:  db,
	}
}

func (p *PaymentGormRepository) GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error) {
	r, err := p.dao.GetPayment(ctx, bizTradeNO)
	return p.toDomain(r), err
}

func (p *PaymentGormRepository) FindExpiredPayment(ctx context.Context, limit, offset int, t time.Time) ([]domain.Payment, error) {
	pmts, err := p.dao.FindExpiredPayment(ctx, limit, offset, t)
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[dao.Payment, domain.Payment](pmts, func(idx int, src dao.Payment) domain.Payment {
		return p.toDomain(src)
	}), nil
}

func (p *PaymentGormRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(pmt))
}

func (p *PaymentGormRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNO, pmt.TxnID, pmt.Status)
}

func (p *PaymentGormRepository) Transaction(ctx context.Context, cb func(pmt *PaymentGormRepository, msg *LocalMsgGormRepository) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return cb(NewPaymentGormRepository(tx), NewLocalMsgGormRepository(tx))
	})
}

func (p *PaymentGormRepository) toDomain(pmt dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Currency: pmt.Currency,
			Total:    pmt.Amt,
		},
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}

func (p *PaymentGormRepository) toEntity(pmt domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		BizTradeNO:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatusInit.AsUint8(),
	}
}
