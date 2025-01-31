package events

import "context"

//go:generate mockgen -source=./types.go -destination=./mocks/producer.mock.go -package=eventsmocks Producer
type Producer interface {
	ProduceSyncEvent(ctx context.Context, data BizTags) error
}

type BizTags struct {
	Tags  []string `json:"tags"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Uid   int64    `json:"uid"`
}

type SyncDataEvent struct {
	IndexName string
	DocId     string
	Data      string
}
