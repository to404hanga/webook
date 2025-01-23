package repository

import (
	"context"
	"webook/account/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/account.mock.go -package=repomocks AccountRepository
type AccountRepository interface {
	AddCredit(ctx context.Context, c domain.Credit) error
}
