package cache

import (
	"context"
	"webook/account/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/account.mock.go -package=cachemocks AccountCache
type AccountCache interface {
	SetUnique(ctx context.Context, cr domain.Credit) error
	GetUnique(ctx context.Context, cr domain.Credit) error
}
