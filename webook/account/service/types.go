package service

import (
	"context"
	"webook/account/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/account.mock.go -package=svcmocks AccountService
type AccountService interface {
	Credit(ctx context.Context, cr domain.Credit) error
}
