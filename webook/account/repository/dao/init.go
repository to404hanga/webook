package dao

import (
	"time"
	"webook/account/domain"

	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	err := db.AutoMigrate(&Account{}, &AccountActivity{})
	if err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	_ = db.Create(&Account{
		Type:       domain.AccountTypeSystem.AsUint8(),
		Currency:   "CNY",
		CreateTime: now,
		UpdateTime: now,
	}).Error
	return nil
}
