package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	dao "webook/internal/repository/dao/article"
	"webook/pkg/logger"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Update(ctx context.Context, article domain.Article) error
	Create(ctx context.Context, article domain.Article) (int64, error)
	ListPub(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]domain.Article, error)
	SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDAO
	cache    cache.ArticleCache
	userRepo UserRepository
	l        logger.Logger
}

var _ ArticleRepository = (*CachedArticleRepository)(nil)

func NewArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, userRepo UserRepository, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		l:        l,
	}
}

func (c *CachedArticleRepository) Cache() cache.ArticleCache {
	return c.cache
}

func (c *CachedArticleRepository) ListPub(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error) {
	res, err := c.dao.ListPub(ctx, start, limit, offset)
	if err != nil {
		return nil, err
	}
	articles := make([]domain.Article, 0, len(res))
	for _, a := range res {
		articles = append(articles, c.toDomain(dao.Article(a)))
	}
	return articles, nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById_SetCacheAsync(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res.Author.Name = author.Nickname
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			c.l.Warn("ArticleRepo GetPubById 设置缓存失败", logger.Error(er))
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(art)
	go func() {
		er := c.cache.Set(ctx, res)
		if er != nil {
			c.l.Warn("ArticleRepo GetById 设置缓存失败", logger.Error(er))
		}
	}()
	return res, nil
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]domain.Article, error) {
	if limit <= 100 && offset == 0 {
		res, err := c.cache.GetFirstPage(ctx, userId)
		if err == nil {
			return res[:limit], nil
		} else {
			c.l.Warn("ArticleRepo GetByAuthor 未命中缓存", logger.Error(err))
		}
	}
	res, err := c.dao.GetByAuthor(ctx, userId, limit, offset)
	if err != nil {
		return nil, err
	}
	articles := make([]domain.Article, 0, len(res))
	for _, a := range res {
		articles = append(articles, c.toDomain(a))
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if limit <= 100 && offset == 0 {
			if err = c.cache.SetFirstPage(ctx, userId, articles); err != nil {
				c.l.Warn("ArticleRepo GetByAuthor SetFirstPage 设置缓存失败", logger.Error(err))
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, articles)
	}()
	return articles, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, userId, id, domain.ArticleStatus(status.ToUint8()))
	if err != nil {
		return err
	}
	err = c.cache.DelFirstPage(ctx, userId)
	if err != nil {
		c.l.Warn("ArticleRepo SyncStatus DelFirstPage 删除缓存失败", logger.Error(err))
	}
	return nil
}

func (c *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(article))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, article.Author.Id)
		if er != nil {
			c.l.Warn("ArticleRepo Sync DelFirstPage 删除缓存失败", logger.Error(er))
		}
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		user, er := c.userRepo.FindById(ctx, article.Author.Id)
		if er != nil {
			c.l.Warn("ArticleRepo Sync 获取用户信息失败", logger.Error(er))
			return
		}
		article.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		er = c.cache.SetPub(ctx, article)
		if er != nil {
			c.l.Warn("ArticleRepo Sync SetPub 设置缓存失败", logger.Error(er))
		}
	}()
	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(article))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, article.Author.Id)
		if er != nil {
			c.l.Warn("ArticleRepo Update DelFirstPage 删除缓存失败", logger.Error(er))
		}
	}
	return err
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(article))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, article.Author.Id)
		if er != nil {
			c.l.Warn("ArticleRepo Create DelFirstPage 删除缓存失败", logger.Error(er))
		}
	}
	return id, err
}

func (c *CachedArticleRepository) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) toDomain(article dao.Article) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Author: domain.Author{
			Id: article.AuthorId,
		},
		Status: domain.ArticleStatus(article.Status),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, articles []domain.Article) {
	const size = 1 << 20
	if len(articles) > 0 && len(articles[0].Content) < size {
		err := c.cache.Set(ctx, articles[0])
		if err != nil {
			c.l.Warn("ArticleRepo preCache 设置缓存失败", logger.Error(err))
		}
	}
}
