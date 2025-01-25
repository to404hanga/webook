package service

import (
	"context"
	"webook/comment/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/comment.mock.go -package=svcmocks CommentService
type CommentService interface {
	GetCommentList(ctx context.Context, biz string, bizId, minId int64, limit int) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	CreateComment(ctx context.Context, comment domain.Comment) error
	GetMoreReplies(ctx context.Context, rootId, maxId int64, limit int) ([]domain.Comment, error)
}
