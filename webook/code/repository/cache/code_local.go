package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/redis/go-redis/v9"
)

var ErrKeyNotExist = redis.Nil

type LocalCodeCache struct {
	// TODO 自己实现一个 lru.Cache
	cache      *lru.Cache
	lock       sync.Mutex
	expiration time.Duration
}

var _ CodeCache = (*LocalCodeCache)(nil)

func NewLocalCodeCache(c *lru.Cache, expiration time.Duration) CodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	codeKey := key(biz, phone)

	now := time.Now()
	val, ok := l.cache.Get(codeKey)
	if !ok {
		// 没有验证码
		l.cache.Add(codeKey, codeItem{
			code:   code,
			cnt:    3,
			expire: now.Add(l.expiration),
		})
		return nil
	}
	item, ok := val.(codeItem)
	if !ok {
		return errors.New("系统错误")
	}
	if item.expire.Sub(now) > 9*time.Minute {
		return ErrCodeSendTooMany
	}
	// 重发
	l.cache.Add(codeKey, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(l.expiration),
	})
	return nil
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	codeKey := key(biz, phone)
	val, ok := l.cache.Get(codeKey)
	if !ok {
		// 没发验证码
		return false, ErrKeyNotExist
	}
	item, ok := val.(codeItem)
	if !ok {
		return false, errors.New("系统错误")
	}
	if item.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	item.cnt--
	return item.code == code, nil
}

type codeItem struct {
	code   string
	cnt    int // 可验证次数
	expire time.Time
}
