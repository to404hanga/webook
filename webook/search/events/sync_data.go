package events

import (
	"context"
	"time"
	"webook/search/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

type SyncDataEvent struct {
	IndexName string
	DocId     string
	Data      string
}

type SyncDataEventConsumer struct {
	svc    service.SyncService
	client sarama.Client
	l      logger.Logger
}

var _ Consumer = (*SyncDataEventConsumer)(nil)

func NewSyncDataEventConsumer(svc service.SyncService, client sarama.Client, l logger.Logger) *SyncDataEventConsumer {
	return &SyncDataEventConsumer{
		svc:    svc,
		client: client,
		l:      l,
	}
}

func (s *SyncDataEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data", s.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(), []string{"sync_search_data"}, saramax.NewHandler(s.l, s.Consume))
		if err != nil {
			s.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (s *SyncDataEventConsumer) Consume(sg *sarama.ConsumerMessage, evt SyncDataEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.svc.InputAny(ctx, evt.IndexName, evt.DocId, evt.Data)
}
