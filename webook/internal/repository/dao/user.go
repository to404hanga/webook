package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, user User) error {
	current := time.Now().UnixMilli()
	user.CreateTime = current
	user.UpdateTime = current
	err := dao.db.WithContext(ctx).Create(&user).Error
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		const duplicateError uint16 = 1062
		if mysqlError.Number == duplicateError {
			return ErrDuplicateEmail
		}
	}
	return err
}

type User struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Email      string `gorm:"unique"`
	Password   string
	CreateTime int64
	UpdateTime int64
	Nickname   string `gorm:"type=varchar(128)"`
	Birthday   int64
	AboutMe    string `gorm:"type=varchar(4096)"`
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (dao *UserDAO) UpdateById(ctx context.Context, user User) error {
	return dao.db.WithContext(ctx).Model(&user).Where("id = ?", user.Id).Updates(map[string]interface{}{
		"update_time": time.Now().UnixMilli(),
		"nickname":    user.Nickname,
		"birthday":    user.Birthday,
		"about_me":    user.AboutMe,
	}).Error
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}
