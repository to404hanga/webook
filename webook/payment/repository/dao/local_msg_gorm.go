package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type LocalMsgGormDAO struct {
	db *gorm.DB
}

var _ LocalMsgDAO = (*LocalMsgGormDAO)(nil)

func NewLocalMsgGormDAO(db *gorm.DB) *LocalMsgGormDAO {
	return &LocalMsgGormDAO{db: db}
}

func (dao *LocalMsgGormDAO) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Msg{}).Where("id = ?", id).Updates(map[string]any{
		"status":      status,
		"update_time": now,
	}).Error
}

func (dao *LocalMsgGormDAO) AddMsg(ctx context.Context, msg Msg) (int64, error) {
	now := time.Now().UnixMilli()
	msg.CreateTime = now
	msg.UpdateTime = now
	err := dao.db.WithContext(ctx).Create(&msg).Error
	return msg.Id, err
}

func (dao *LocalMsgGormDAO) FindInitMsg(ctx context.Context, limit, offset int) ([]Msg, error) {
	var msgs []Msg
	err := dao.db.WithContext(ctx).Where("status = ?", MsgStatusInit).Offset(offset).Limit(limit).Find(&msgs).Error
	return msgs, err
}
