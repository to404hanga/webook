package startup

import (
	dao2 "webook/interactive/repository/dao"
	"webook/internal/repository/dao"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:123456@tcp(localhost:3307)/webook"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
