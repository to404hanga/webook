package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(ctx context.Context, user domain.User) error {
	return repo.dao.Insert(ctx, dao.User{
		Email:    user.Email,
		Password: user.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(user), nil
}

func (repo *UserRepository) toDomain(user dao.User) domain.User {
	return domain.User{
		Id:         user.Id,
		Email:      user.Email,
		Password:   user.Password,
		CreateTime: time.UnixMilli(user.CreateTime),
		Nickname:   user.Nickname,
		Birthday:   time.UnixMilli(user.Birthday),
		AboutMe:    user.AboutMe,
	}
}

func (repo *UserRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		Nickname: user.Nickname,
		Birthday: user.Birthday.UnixMilli(),
		AboutMe:  user.AboutMe,
	}
}

func (repo *UserRepository) UpdateById(ctx context.Context, user domain.User) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(user))
}

func (repo *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(user), nil
}
