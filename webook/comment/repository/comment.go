package repository

import (
	"context"
	"database/sql"
	"time"
	"webook/comment/domain"
	"webook/comment/repository/dao"

	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/stl/transform"
	"github.com/to404hanga/pkg404/stl/vector"
	"golang.org/x/sync/errgroup"
)

type CachedCommentRepository struct {
	dao dao.CommentDAO
	l   logger.Logger
}

var _ CommentRepository = (*CachedCommentRepository)(nil)

func NewCachedCommentRepository(dao dao.CommentDAO, l logger.Logger) CommentRepository {
	return &CachedCommentRepository{
		dao: dao,
		l:   l,
	}
}

func (c *CachedCommentRepository) FindByBiz(ctx context.Context, biz string, bizId, minId int64, limit int) ([]domain.Comment, error) {
	cs, err := c.dao.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(cs))
	vec := vector.NewSliceVectorFromSlice(cs...)
	var eg errgroup.Group
	downgraded := ctx.Value("downgraded") == "true"
	vec.ForEach(func(val dao.Comment) {
		cmt := c.toDomain(val)
		res = append(res, cmt)
		if !downgraded {
			eg.Go(func() error {
				// 只展示 3 条
				cmt.Children = make([]domain.Comment, 0, 3)
				rs, err := c.dao.FindRepliesByParentId(ctx, val.Id, 3, 0)
				if err != nil {
					c.l.Error("查询子评论失败", logger.Error(err))
					return nil
				}
				for _, r := range rs {
					cmt.Children = append(cmt.Children, c.toDomain(r))
				}
				return nil
			})
		}
	})
	return res, eg.Wait()
}

func (c *CachedCommentRepository) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, dao.Comment{
		Id: comment.Id,
	})
}

func (c *CachedCommentRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *CachedCommentRepository) GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error) {
	vals, err := c.dao.FindOneByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	cs := transform.SliceFromSlice[dao.Comment, domain.Comment](vals, func(cmt dao.Comment) domain.Comment {
		return c.toDomain(cmt)
	})
	return cs, nil
}

func (c *CachedCommentRepository) GetMoreReplies(ctx context.Context, rootId, maxId int64, limit int) ([]domain.Comment, error) {
	cs, err := c.dao.FindRepliesByRootId(ctx, rootId, maxId, limit)
	if err != nil {
		return nil, err
	}
	res := transform.SliceFromSlice[dao.Comment, domain.Comment](cs, func(cmt dao.Comment) domain.Comment {
		return c.toDomain(cmt)
	})
	return res, nil
}

func (c *CachedCommentRepository) toDomain(cmt dao.Comment) domain.Comment {
	val := domain.Comment{
		Id: cmt.Id,
		Commentator: domain.User{
			Id: cmt.Uid,
		},
		Biz:        cmt.Biz,
		BizId:      cmt.BizId,
		Content:    cmt.Content,
		CreateTime: time.UnixMilli(cmt.CreateTime),
		UpdateTime: time.UnixMilli(cmt.UpdateTime),
	}
	if cmt.ParentId.Valid {
		val.ParentComment = &domain.Comment{
			Id: cmt.ParentId.Int64,
		}
	}
	if cmt.RootId.Valid {
		val.RootComment = &domain.Comment{
			Id: cmt.RootId.Int64,
		}
	}
	return val
}

func (c *CachedCommentRepository) toEntity(cmt domain.Comment) dao.Comment {
	now := time.Now().UnixMilli()
	val := dao.Comment{
		Id:         cmt.Id,
		Uid:        cmt.Commentator.Id,
		Biz:        cmt.Biz,
		BizId:      cmt.BizId,
		Content:    cmt.Content,
		CreateTime: now,
		UpdateTime: now,
	}
	if cmt.RootComment != nil {
		val.RootId = sql.NullInt64{
			Valid: true,
			Int64: cmt.RootComment.Id,
		}
	}
	if cmt.ParentComment != nil {
		val.ParentId = sql.NullInt64{
			Valid: true,
			Int64: cmt.ParentComment.Id,
		}
	}
	return val
}
