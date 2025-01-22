package repository

import (
	"context"
	"time"
	"webook/payment/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/payment.mock.go -package=repomocks PaymentRepository
type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, limit, offset int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeNO string) (domain.Payment, error)
}
