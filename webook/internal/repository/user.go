package repository

import (
	"context"
	"database/sql"
	"log"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateById(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindById_SetCacheAsync(ctx context.Context, id int64) (domain.User, error)
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
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
		Nickname:   user.Nickname,
		Birthday:   time.UnixMilli(user.Birthday),
		AboutMe:    user.AboutMe,
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
		Nickname: user.Nickname,
		Birthday: user.Birthday.UnixMilli(),
		AboutMe:  user.AboutMe,
	}
}

func (repo *CachedUserRepository) UpdateById(ctx context.Context, user domain.User) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(user))
}

// 使用 mock 进行单元测试时只可使用同步的写法
func (repo *CachedUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, id)
	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist:
		user, err := repo.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(user)

		// // 异步刷新缓存，提高性能
		// go func() {
		// 	err = repo.cache.Set(ctx, du)
		// 	if err != nil {
		// 		// 可能导致缓存击穿
		// 		log.Println(err)
		// 	}
		// }()

		err = repo.cache.Set(ctx, du)
		if err != nil {
			// 可能导致缓存击穿
			log.Println(err)
		}

		return du, nil
	default:
		// redis 异常，类似降级
		return domain.User{}, err
	}
}

// 异步刷新缓存的 FindById 方法
func (repo *CachedUserRepository) FindById_SetCacheAsync(ctx context.Context, id int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, id)
	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist:
		user, err := repo.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(user)

		// 异步刷新缓存，提高性能
		go func() {
			err = repo.cache.Set(ctx, du)
			if err != nil {
				// 可能导致缓存击穿
				log.Println(err)
			}
		}()

		return du, nil
	default:
		// redis 异常，类似降级
		return domain.User{}, err
	}
}
