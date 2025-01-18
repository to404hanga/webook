package dao

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
	ErrUnknownPattern = errors.New("未知的双写模式")
)
