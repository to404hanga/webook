package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"webook/user/domain"
	"webook/user/repository/cache"
	"webook/user/repository/dao"

	"github.com/to404hanga/pkg404/logger"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

//go:generate mockgen -source=./user.go -package=repomocks -destination=./mocks/user.mock.go UserRepository
type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateById(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindById_SetCacheAsync(ctx context.Context, id int64) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
	l     logger.Logger
}

func NewCachedUserRepository(dao dao.UserDAO, cache cache.UserCache, l logger.Logger) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

func (repo *CachedUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	user, err := repo.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(user), nil
}

func (repo *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(user))
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(user), nil
}

func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(user), nil
}

func (repo *CachedUserRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:         user.Id,
		Email:      user.Email.String,
		Phone:      user.Phone.String,
		Password:   user.Password,
		CreateTime: time.UnixMilli(user.CreateTime),
		Nickname:   user.Nickname.String,
		Birthday:   time.UnixMilli(user.Birthday.Int64),
		AboutMe:    user.AboutMe.String,
		WechatInfo: domain.WechatInfo{
			OpenId:  user.WechatOpenId.String,
			UnionId: user.WechatUnionId.String,
		},
	}
}

func (repo *CachedUserRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		Nickname: sql.NullString{
			String: user.Nickname,
			Valid:  user.Nickname != "",
		},
		Birthday: sql.NullInt64{
			Int64: user.Birthday.UnixMilli(),
			Valid: user.Birthday.UnixMilli() != 0,
		},
		WechatOpenId: sql.NullString{
			String: user.WechatInfo.OpenId,
			Valid:  user.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: user.WechatInfo.UnionId,
			Valid:  user.WechatInfo.UnionId != "",
		},
		AboutMe: sql.NullString{
			String: user.AboutMe,
			Valid:  user.AboutMe != "",
		},
	}
}

func (repo *CachedUserRepository) UpdateById(ctx context.Context, user domain.User) error {
	err := repo.dao.UpdateNonZeroFields(ctx, repo.toEntity(user))
	if err != nil {
		return err
	}
	// 延迟双删
	time.AfterFunc(time.Second, func() {
		repo.cache.Del(ctx, user.Id)
	})
	return repo.cache.Del(ctx, user.Id)
}

// 使用 mock 进行单元测试时只可使用同步的写法
func (repo *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, id)
	if err == nil {
		return du, nil
	}

	if ctx.Value("downgrade") == "true" || ctx.Value("limited") == "true" {
		return domain.User{}, errors.New("触发限流或降级，不再查询数据库")
	}

	user, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(user)

	err = repo.cache.Set(ctx, du)
	if err != nil {
		repo.l.Error("写入缓存失败", logger.Error(err))
	}
	return du, nil
}

// 异步刷新缓存的 FindById 方法
func (repo *CachedUserRepository) FindById_SetCacheAsync(ctx context.Context, id int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, id)
	if err == nil {
		return du, nil
	}

	if ctx.Value("downgrade") == "true" || ctx.Value("limited") == "true" {
		return domain.User{}, errors.New("触发限流或降级，不再查询数据库")
	}

	user, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(user)

	go func() {
		err = repo.cache.Set(ctx, du)
		if err != nil {
			repo.l.Error("写入缓存失败", logger.Error(err))
		}
	}()

	return du, nil
}
