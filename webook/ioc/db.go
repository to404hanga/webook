package ioc

import (
	"webook/internal/repository/dao"
	"webook/pkg/logger"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	db, err := gorm.Open(mysql.Open(viper.GetString("db.dsn")), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold: 0, // 关闭慢查询
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, val ...interface{}) {
	g(msg, logger.Any("args", val))
}
