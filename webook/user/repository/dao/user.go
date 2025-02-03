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
	ErrUserDuplicate = errors.New("用户邮箱或手机号冲突")
	ErrDataNotFound  = gorm.ErrRecordNotFound
)

//go:generate mockgen -source=./user.go -package=daomocks -destination=./mocks/user.mock.go UserDAO
type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateNonZeroFields(ctx context.Context, user User) error
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewGormUserDAO(db *gorm.DB) UserDAO {
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
			return ErrUserDuplicate
		}
	}
	return err
}

type User struct {
	Id            int64          `gorm:"primaryKey,autoIncrement"`
	Email         sql.NullString `gorm:"unique"`
	Password      string
	Phone         sql.NullString `gorm:"unique"`
	Birthday      sql.NullInt64  // yyyy-MM-dd
	Nickname      sql.NullString `gorm:"type=varchar(50)"`
	AboutMe       sql.NullString `gorm:"type=varchar(1024)"`
	WechatOpenId  sql.NullString `gorm:"type:varchar(255);unique"`
	WechatUnionId sql.NullString `gorm:"type:varchar(255)"`
	CreateTime    int64
	UpdateTime    int64
}

func (dao *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (dao *GormUserDAO) UpdateNonZeroFields(ctx context.Context, user User) error {
	return dao.db.WithContext(ctx).Model(&user).Updates(&user).Error
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
