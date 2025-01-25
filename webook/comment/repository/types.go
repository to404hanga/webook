package repository

import (
	"context"
	"webook/comment/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/comment.mock.go -package=repomocks CommentRepository
type CommentRepository interface {
	FindByBiz(ctx context.Context, biz string, bizId, minId int64, limit int) ([]domain.Comment, error) // FindByBiz 根据 id 倒叙查找，返回每个评论的三个直接回复
	DeleteComment(ctx context.Context, comment domain.Comment) error                                    // DeleteComment 删除评论及其子评论
	CreateComment(ctx context.Context, comment domain.Comment) error
	GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rootId, maxId int64, limit int) ([]domain.Comment, error)
}
