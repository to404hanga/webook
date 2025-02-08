package repository

import (
	"context"
	"time"
	"webook/article/domain"
	"webook/article/repository/cache"
	"webook/article/repository/dao"

	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/stl/transform"
)

//go:generate mockgen -source=./article.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	List(ctx context.Context, author int64, limit, offset int) ([]domain.Article, error)
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, limit, offset int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDAO
	cache cache.ArticleCache
	l     logger.Logger
}

var _ ArticleRepository = (*CachedArticleRepository)(nil)

func NewArticleRepository(dao dao.ArticleDAO, cache cache.ArticleCache, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		cache: cache,
		l:     l,
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
		articles = append(articles, c.ToDomain(dao.Article(a)))
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
	res = domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.cache.SetPub(ctx, res)
		if er != nil {
			c.l.Warn("缓存已发表失败", logger.Error(er), logger.Int64("aid", res.Id))
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
	return c.ToDomain(art), nil
}

func (c *CachedArticleRepository) List(ctx context.Context, author int64, limit, offset int) ([]domain.Article, error) {
	if offset == 0 && limit == 100 {
		data, err := c.cache.GetFirstPage(ctx, author)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
		if err != cache.ErrKeyNotExist {
			c.l.Error("查询缓存文章失败", logger.Error(err), logger.Int64("author", author))
		}
	}
	articles, err := c.dao.GetByAuthor(ctx, author, limit, offset)
	if err != nil {
		return nil, err
	}
	res := transform.SliceFromSlice[dao.Article, domain.Article](articles, func(idx int, a dao.Article) domain.Article {
		return c.ToDomain(a)
	})

	go func() {
		c.preCache(ctx, res)
	}()

	err = c.cache.SetFirstPage(ctx, author, res)
	if err != nil {
		c.l.Error("刷新第一页文章的缓存失败", logger.Error(err), logger.Int64("author", author))
	}
	return res, nil
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error {
	return c.dao.SyncStatus(ctx, userId, id, domain.ArticleStatus(status.ToUint8()))
}

func (c *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(article))
	if err != nil {
		return 0, err
	}

	go func() {
		author := article.Author.Id
		err = c.cache.DelFirstPage(ctx, author)
		if err != nil {
			c.l.Error("删除第一页缓存失败", logger.Error(err), logger.Int64("author", author))
		}
		err = c.cache.SetPub(ctx, article)
		if err != nil {
			c.l.Error("提前设置缓存失败", logger.Error(err), logger.Int64("author", author))
		}
	}()

	return id, err
}

func (c *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(article))
	if err != nil {
		return err
	}

	go func() {
		author := article.Author.Id
		err = c.cache.DelFirstPage(ctx, author)
		if err != nil {
			c.l.Error("删除缓存失败", logger.Error(err), logger.Int64("author", author))
		}
	}()

	return nil
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(article))
	if err != nil {
		return 0, err
	}

	go func() {
		author := article.Author.Id
		err = c.cache.DelFirstPage(ctx, author)
		if err != nil {
			c.l.Error("删除缓存失败", logger.Error(err), logger.Int64("author", author))
		}
	}()

	return id, nil
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

func (c *CachedArticleRepository) ToDomain(article dao.Article) domain.Article {
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
	if len(articles) > 0 && len(articles[0].Content) <= size {
		err := c.cache.Set(ctx, articles[0])
		if err != nil {
			c.l.Error("提前缓存失败", logger.Error(err))
		}
	}
}
