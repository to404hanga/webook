package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateById(ctx context.Context, user User) error
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{
		db: db,
	}
}

func (dao *GormUserDAO) FindByWechat(ctx context.Context, openId string) (User, error) {

	var user User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&user).Error
	return user, err
}

func (dao *GormUserDAO) Insert(ctx context.Context, user User) error {
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
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string

	Nickname string `gorm:"type=varchar(128)"`
	Birthday int64  // yyyy-MM-dd
	AboutMe  string `gorm:"type=varchar(4096)"`

	Phone sql.NullString `gorm:"unique"`

	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString

	CreateTime int64
	UpdateTime int64
}

func (dao *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (dao *GormUserDAO) UpdateById(ctx context.Context, user User) error {
	return dao.db.WithContext(ctx).Model(&user).Where("id = ?", user.Id).Updates(map[string]interface{}{
		"update_time": time.Now().UnixMilli(),
		"nickname":    user.Nickname,
		"birthday":    user.Birthday,
		"about_me":    user.AboutMe,
	}).Error
}

func (dao *GormUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (dao *GormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return user, err
}
