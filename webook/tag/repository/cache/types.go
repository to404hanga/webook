package cache

import (
	"context"
	"webook/tag/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/tag.mock.go -package=cachemocks TagCache
type TagCache interface {
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	Append(ctx context.Context, uid int64, tags ...domain.Tag) error
	DelTags(ctx context.Context, uid int64) error
}
