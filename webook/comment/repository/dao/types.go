package dao

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=./types.go -package=daomocks -destination=./mocks/comment.mock.go CommentDAO
type CommentDAO interface {
	Insert(ctx context.Context, u Comment) error
	FindByBiz(ctx context.Context, biz string, bizId, minId int64, limit int) ([]Comment, error) // FindByBiz 只查找一级评论
	FindCommentList(ctx context.Context, u Comment) ([]Comment, error)                           // FindCommentList Comment.Id 为 0 时 获取一级评论，否则获取对应的评论以及其所有回复
	FindRepliesByParentId(ctx context.Context, parentId int64, limit, offset int) ([]Comment, error)
	Delete(ctx context.Context, u Comment) error // Delete 删除本节点和其对应的所有子节点
	FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error)
	FindRepliesByRootId(ctx context.Context, rootId, id int64, limit int) ([]Comment, error)
}

type Comment struct {
	Id            int64         `gorm:"column:id;primaryKey" json:"id"`
	Uid           int64         `gorm:"column:uid;index" json:"uid"`
	Biz           string        `gorm:"column:biz;index:biz_type_id" json:"biz"`
	BizId         int64         `gorm:"column:biz_id;index:biz_type_id" json:"bizId"`
	RootId        sql.NullInt64 `gorm:"column:root_id;index" json:"rootId"`
	ParentId      sql.NullInt64 `gorm:"column:parent_id;index" json:"parentId"`
	ParentComment *Comment      `gorm:"ForeignKey:ParentId;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	Content       string        `gorm:"type:text;column:content" json:"content"`
	CreateTime    int64         `gorm:"column:create_time" json:"createTime"`
	UpdateTime    int64         `gorm:"column:update_time" json:"updateTime"`
}

func (Comment) TableName() string {
	return "comment"
}
