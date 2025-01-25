package dao

import (
	"context"

	"gorm.io/gorm"
)

type CommentGormDAO struct {
	db *gorm.DB
}

var _ CommentDAO = (*CommentGormDAO)(nil)

func NewCommentGormDAO(db *gorm.DB) CommentDAO {
	return &CommentGormDAO{db: db}
}

func (c *CommentGormDAO) Insert(ctx context.Context, u Comment) error {
	return c.db.WithContext(ctx).Create(u).Error
}

func (c *CommentGormDAO) FindByBiz(ctx context.Context, biz string, bizId, minId int64, limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND id < ? AND parent_id IS NULL", biz, bizId, minId).Limit(limit).Find(&res).Error
	return res, err
}

func (c *CommentGormDAO) FindCommentList(ctx context.Context, u Comment) ([]Comment, error) {
	var res []Comment
	builder := c.db.WithContext(ctx)
	if u.Id == 0 {
		builder = builder.Where("biz = ? AND biz_id = ? AND root_id IS NULL", u.Biz, u.BizId)
	} else {
		builder = builder.Where("root_id = ? OR id = ?", u.Id, u.Id)
	}
	err := builder.Find(&res).Error
	return res, err
}

func (c *CommentGormDAO) FindRepliesByParentId(ctx context.Context, parentId int64, limit, offset int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("parent_id = ?", parentId).Order("id DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (c *CommentGormDAO) Delete(ctx context.Context, u Comment) error {
	return c.db.WithContext(ctx).Delete(&Comment{
		Id: u.Id,
	}).Error
}

func (c *CommentGormDAO) FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

func (c *CommentGormDAO) FindRepliesByRootId(ctx context.Context, rootId, id int64, limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("root_id = ? AND id > ?", rootId, id).Order("id ASC").Limit(int(limit)).Find(&res).Error
	return res, err
}
