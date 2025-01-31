package repository

import (
	"context"
	"webook/tag/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/tag.mock.go -package=repomocks TagRepository
type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
	PreloadUserTags(ctx context.Context) error
}
