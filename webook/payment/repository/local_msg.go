package repository

import (
	"context"
	"time"
	"webook/payment/domain"
	"webook/payment/repository/dao"

	"github.com/to404hanga/pkg404/stl/transform"
	"gorm.io/gorm"
)

type LocalMsgGormRepository struct {
	dao *dao.LocalMsgGormDAO
}

var _ LocalMsgRepository = (*LocalMsgGormRepository)(nil)

func NewLocalMsgGormRepository(db *gorm.DB) *LocalMsgGormRepository {
	return &LocalMsgGormRepository{
		dao: dao.NewLocalMsgGormDAO(db),
	}
}

func (l *LocalMsgGormRepository) FindInitMsg(ctx context.Context, limit, offset int) ([]domain.Msg, error) {
	msgs, err := l.dao.FindInitMsg(ctx, limit, offset)
	return transform.SliceFromSlice(msgs, func(idx int, src dao.Msg) domain.Msg {
		return domain.Msg{
			Id:         src.Id,
			Content:    src.Content,
			CreateTime: time.UnixMilli(src.CreateTime),
		}
	}), err
}

func (l *LocalMsgGormRepository) MarkFailed(ctx context.Context, id int64) error {
	return l.dao.UpdateStatus(ctx, id, dao.MsgStatusFailed)
}

func (l *LocalMsgGormRepository) MarkSuccess(ctx context.Context, id int64) error {
	return l.dao.UpdateStatus(ctx, id, dao.MsgStatusSuccess)
}

func (l *LocalMsgGormRepository) AddMsg(ctx context.Context, content string) (int64, error) {
	return l.dao.AddMsg(ctx, dao.Msg{
		Content: content,
		Status:  dao.MsgStatusInit,
	})
}
