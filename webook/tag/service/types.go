package service

import (
	"context"
	"webook/tag/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/tag.mock.go -package=svcmocks TagService
type TagService interface {
	CreateTag(ctx context.Context, uid int64, name string) (int64, error)
	AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
}
