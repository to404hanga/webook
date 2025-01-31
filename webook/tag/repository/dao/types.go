package dao

import "context"

type Tag struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type=varchar(4096)"`
	Uid        int64  `gorm:"index"`
	CreateTime int64
	UpdateTime int64
}

type TagBiz struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"index:biz_type_id"`
	Biz        string `gorm:"index:biz_type_id"`
	Uid        int64  `gorm:"index"`
	Tid        int64
	Tag        *Tag  `gorm:"ForeignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	CreateTime int64 `bson:"create_time,omitempty"`
	UpdateTime int64 `bson:"update_time,omitempty"`
}

//go:generate mockgen -source=./types.go -destination=./mocks/tag.mock.go -package=daomocks TagDAO
type TagDAO interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, limit, offset int) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
}
