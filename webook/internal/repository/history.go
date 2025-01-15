package repository

import (
	"context"
	"webook/internal/domain"
)

//go:generate mockgen -source=./history.go -package=repomocks -destination=./mocks/history.mock.go HistoryRecordRepository
type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
