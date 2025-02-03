package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/to404hanga/pkg404/logger"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送验证码太频繁")
	ErrUnknownForCode    = errors.New("发送验证码时遇到未知错误")
	ErrCodeVerifyTooMany = errors.New("验证次数过多")
)

//go:generate mockgen -source=./code.go -package=cachemocks -destination=./mocks/code.mock.go CodeCache
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type RedisCodeCache struct {
	cmd redis.Cmdable
	l   logger.Logger
}

func NewRedisCodeCache(cmd redis.Cmdable, l logger.Logger) CodeCache {
	return &RedisCodeCache{
		cmd: cmd,
		l:   l,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCodeSendTooMany
	default:
		c.l.Error("发送验证码时遇到未知错误", logger.Int("错误代码", res))
		return ErrUnknownForCode
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return false, nil
	}
}

func key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
