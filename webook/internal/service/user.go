package service

import (
	"context"
	"errors"
	"webook/internal/domain"
	"webook/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateUser         = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户名或密码错误")
)

type UserService interface {
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
	SignUp(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	EditNonSensitive(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, userId int64) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	user, err := svc.repo.FindByWechat(ctx, info.OpenID)
	if err != repository.ErrUserNotFound {
		return user, err
	}
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: info,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	// 要么是用户已存在，要么是用户创建成功
	// ! 主从延迟，理论上强制走主库
	return svc.repo.FindByWechat(ctx, info.OpenID)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 认为，大部分用户是已存在的用户
	user, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return user, err
	}
	// 如果用户不存在，则创建一个新用户
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	// 要么是用户已存在，要么是用户创建成功
	// ! 主从延迟，理论上强制走主库
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) SignUp(ctx context.Context, user domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return svc.repo.Create(ctx, user)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return domain.User{}, ErrInvalidUserOrPassword
		}
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, nil
}

// 编辑非敏感数据
func (svc *userService) EditNonSensitive(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateById(ctx, user)
}

func (svc *userService) Profile(ctx context.Context, userId int64) (domain.User, error) {
	return svc.repo.FindById(ctx, userId)
}
