package repository

import (
	"context"
	"webook/search/repository/dao"
)

type anyRepository struct {
	dao dao.AnyDAO
}

var _ AnyRepository = (*anyRepository)(nil)

func NewAnyRepository(dao dao.AnyDAO) AnyRepository {
	return &anyRepository{
		dao: dao,
	}
}

func (a *anyRepository) Input(ctx context.Context, index, docId, data string) error {
	return a.dao.Input(ctx, index, docId, data)
}

func (a *anyRepository) Delete(ctx context.Context, index, docId string) error {
	return a.dao.Delete(ctx, index, docId)
}
