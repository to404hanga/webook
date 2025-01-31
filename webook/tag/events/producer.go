package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type SaramaSyncProducer struct {
	client sarama.SyncProducer
}

var _ Producer = (*SaramaSyncProducer)(nil)

func NewSaramaSyncProducer(client sarama.Client) (*SaramaSyncProducer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaSyncProducer{client: p}, nil
}

func (s *SaramaSyncProducer) ProduceSyncEvent(ctx context.Context, tags BizTags) error {
	val, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	evt := SyncDataEvent{
		IndexName: "tags_index",
		DocId:     fmt.Sprintf("%d_%s_%d", tags.Uid, tags.Biz, tags.BizId),
		Data:      string(val),
	}
	val, err = json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = s.client.SendMessage(&sarama.ProducerMessage{
		Topic: "sync_search_data",
		Value: sarama.ByteEncoder(val),
	})
	return err
}
