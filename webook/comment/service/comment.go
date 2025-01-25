package service

import (
	"context"
	"webook/comment/domain"
	"webook/comment/repository"
)

type commentService struct {
	repo repository.CommentRepository
}

var _ CommentService = (*commentService)(nil)

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{
		repo: repo,
	}
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId, minId int64, limit int) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.repo.DeleteComment(ctx, domain.Comment{
		Id: id,
	})
}

func (c *commentService) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.repo.CreateComment(ctx, comment)
}

func (c *commentService) GetMoreReplies(ctx context.Context, rootId, maxId int64, limit int) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rootId, maxId, limit)
}
