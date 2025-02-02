package repository

import (
	"context"
	"webook/sms/domain"
	"webook/sms/repository/dao"

	"github.com/ecodeclub/ekit/sqlx"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

//go:generate mockgen -source=./async.go -package=repomocks -destination=./mocks/async.mock.go AsyncSmsRepository
type AsyncSmsRepository interface {
	Add(ctx context.Context, s domain.AsyncSms) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSmsRepository struct {
	dao dao.AsyncSmsDAO
}

var _ AsyncSmsRepository = (*asyncSmsRepository)(nil)

func NewAsyncSmsRepository(dao dao.AsyncSmsDAO) AsyncSmsRepository {
	return &asyncSmsRepository{
		dao: dao,
	}
}

func (repo *asyncSmsRepository) Add(ctx context.Context, s domain.AsyncSms) error {
	return repo.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   s.TplId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (repo *asyncSmsRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := repo.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}
	return domain.AsyncSms{
		Id:       as.Id,
		TplId:    as.Config.Val.TplId,
		Numbers:  as.Config.Val.Numbers,
		Args:     as.Config.Val.Args,
		RetryMax: as.RetryMax,
	}, nil
}

func (repo *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return repo.dao.MarkSuccess(ctx, id)
	}
	return repo.dao.MarkFailed(ctx, id)
}
