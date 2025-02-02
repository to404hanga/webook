package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/article/domain"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -source=./article.go -package=cachemocks -destination=./mocks/article.mock.go ArticleCache
type ArticleCache interface {
	DelFirstPage(ctx context.Context, userId int64) error
	DelPub(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, article domain.Article) error
	GetFirstPage(ctx context.Context, userId int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, userId int64, articles []domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

var _ ArticleCache = (*ArticleRedisCache)(nil)

const ErrKeyNotExist = redis.Nil

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{client: client}
}

func (c *ArticleRedisCache) DelFirstPage(ctx context.Context, userId int64) error {
	return c.client.Del(ctx, c.firstKey(userId)).Err()
}

func (c *ArticleRedisCache) DelPub(ctx context.Context, id int64) error {
	return c.client.Del(ctx, c.pubKey(id)).Err()
}

func (c *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := c.client.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (c *ArticleRedisCache) Set(ctx context.Context, article domain.Article) error {
	val, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(article.Id), val, 10*time.Minute).Err()
}

func (c *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := c.client.Get(ctx, c.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (c *ArticleRedisCache) SetPub(ctx context.Context, article domain.Article) error {
	val, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.pubKey(article.Id), val, 10*time.Minute).Err()
}

func (c *ArticleRedisCache) GetFirstPage(ctx context.Context, userId int64) ([]domain.Article, error) {
	key := c.firstKey(userId)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (c *ArticleRedisCache) SetFirstPage(ctx context.Context, userId int64, articles []domain.Article) error {
	for i := 0; i < len(articles); i++ {
		articles[i].Content = articles[i].Abstract()
	}
	key := c.firstKey(userId)
	val, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, val, 10*time.Minute).Err()
}

func (c *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func (c *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (c *ArticleRedisCache) firstKey(userId int64) string {
	return fmt.Sprintf("article:first_page:%d", userId)
}
