package repository

import (
	"context"
	"webook/search/domain"
	"webook/search/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
)

type userRepository struct {
	dao dao.UserDAO
}

var _ UserRepository = (*userRepository)(nil)

func NewUserRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{dao: dao}
}

func (u *userRepository) InputUser(ctx context.Context, msg domain.User) error {
	return u.dao.InputUser(ctx, dao.User{
		Id:       msg.Id,
		Email:    msg.Email,
		Nickname: msg.Nickname,
		Phone:    msg.Phone,
	})
}

func (u *userRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	users, err := u.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}
	return transform.SliceFromSlice[dao.User, domain.User](users, func(idx int, user dao.User) domain.User {
		return domain.User{
			Id:       user.Id,
			Email:    user.Email,
			Nickname: user.Nickname,
			Phone:    user.Phone,
		}
	}), nil
}
