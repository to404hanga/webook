package dao

import (
	"context"
	"time"

	"github.com/to404hanga/pkg404/stl/transform"
	"gorm.io/gorm"
)

type GormTagDAO struct {
	db *gorm.DB
}

var _ TagDAO = (*GormTagDAO)(nil)

func NewGormTagDAO(db *gorm.DB) TagDAO {
	return &GormTagDAO{db: db}
}

func (g *GormTagDAO) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := g.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

func (g *GormTagDAO) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.CreateTime = now
	tag.UpdateTime = now
	err := g.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

func (g *GormTagDAO) CreateTagBiz(ctx context.Context, tagBizs []TagBiz) error {
	if len(tagBizs) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for idx := range tagBizs {
		tagBizs[idx].CreateTime = now
		tagBizs[idx].UpdateTime = now
	}
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		first := tagBizs[0]
		err := tx.Model(&TagBiz{}).Delete("uid = ? AND biz = ? AND biz_id = ?", first.Uid, first.Biz, first.BizId).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBizs).Error
	})
}

func (g *GormTagDAO) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := g.db.WithContext(ctx).Where("uid = ?", uid).Find(&res).Error
	return res, err
}

func (g *GormTagDAO) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {
	var tagBizs []TagBiz
	err := g.db.WithContext(ctx).Model(&TagBiz{}).InnerJoins("Tag", g.db.Model(&Tag{})).Where("Tag.uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).Find(&tagBizs).Error
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[TagBiz, Tag](tagBizs, func(idx int, tb TagBiz) Tag {
		return *tb.Tag
	}), nil
}

func (g *GormTagDAO) GetTags(ctx context.Context, limit, offset int) ([]Tag, error) {
	var res []Tag
	err := g.db.WithContext(ctx).Model(&Tag{}).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}
