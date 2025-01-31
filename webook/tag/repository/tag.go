package repository

import (
	"context"
	"time"
	"webook/tag/domain"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"

	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/stl/transform"
)

type CachedTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.Logger
}

var _ TagRepository = (*CachedTagRepository)(nil)

func NewCachedTagRepository(dao dao.TagDAO, cache cache.TagCache, l logger.Logger) TagRepository {
	return &CachedTagRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (c *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := c.dao.CreateTag(ctx, c.toEntity(tag))
	if err != nil {
		return 0, err
	}

	go func() {
		err = c.cache.Append(ctx, tag.Uid, tag)
		if err != nil {
			c.l.Warn("CachedTagRepository CreateTag 加入缓存失败", logger.Error(err), logger.Int64("uid", tag.Uid), logger.Int64("id", tag.Id))
		}
	}()

	return id, nil
}

func (c *CachedTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	// 限流或降级时返回空结果
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		return nil, nil
	}

	tags, err := c.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[dao.Tag, domain.Tag](tags, func(t dao.Tag) domain.Tag {
		return c.toDomain(t)
	}), nil
}

func (c *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := c.cache.GetTags(ctx, uid)
	if err == nil {
		return res, nil
	}

	// 限流或降级时返回空结果
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		return nil, nil
	}

	tags, err := c.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	res = transform.SliceFromSlice[dao.Tag, domain.Tag](tags, func(t dao.Tag) domain.Tag {
		return c.toDomain(t)
	})

	go func() {
		err = c.cache.Append(ctx, uid, res...)
		if err != nil {
			c.l.Warn("CachedTagRepository GetTags 加入缓存失败", logger.Error(err), logger.Int64("uid", uid))
		}
	}()

	return res, nil
}

func (c *CachedTagRepository) BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	return c.dao.CreateTagBiz(ctx, transform.SliceFromSlice[int64, dao.TagBiz](tags, func(i int64) dao.TagBiz {
		return dao.TagBiz{
			Tid:   i,
			BizId: bizId,
			Biz:   biz,
			Uid:   uid,
		}
	}))
}

func (c *CachedTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := c.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[dao.Tag, domain.Tag](tags, func(t dao.Tag) domain.Tag {
		return c.toDomain(t)
	}), nil
}

func (c *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	offset := 0
	const batch = 100
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		tags, err := c.dao.GetTags(dbCtx, batch, offset)
		cancel()
		if err != nil {
			c.l.Error("预加载缓存读取数据库失败", logger.Error(err))
			continue
		}

		for _, tag := range tags {
			rdsCtx, cancel := context.WithTimeout(ctx, time.Second)
			err = c.cache.Append(rdsCtx, tag.Uid, c.toDomain(tag))
			cancel()
			if err != nil {
				c.l.Warn("CachedTagRepository PreloadUserTags 加入缓存失败", logger.Error(err), logger.Int64("uid", tag.Uid), logger.Int64("id", tag.Id))
				continue
			}
		}

		if len(tags) < batch {
			return nil
		}
		offset += batch
	}
}

func (c *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (c *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
