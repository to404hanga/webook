package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webook/search/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

const InteractiveTopic = "sync_interactive"

type InteractiveConsumer struct {
	syncSvc  service.SyncService
	client   sarama.Client
	l        logger.Logger
	handlers map[int64]InteractiveHandler
}

var _ Consumer = (*InteractiveConsumer)(nil)

func NewInteractiveConsumer(client sarama.Client, l logger.Logger, svc service.SyncService) *InteractiveConsumer {
	handlers := map[int64]InteractiveHandler{
		1: &LikeHandler{syncSvc: svc},
		2: &CollectHandler{syncSvc: svc},
		3: &CancelLikeHandler{syncSvc: svc},
		4: &CancelCollectHandler{syncSvc: svc},
	}
	return &InteractiveConsumer{
		syncSvc:  svc,
		client:   client,
		l:        l,
		handlers: handlers,
	}
}

func (i *InteractiveConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_interactive_group1", i.client)
	if err != nil {
		return err
	}

	go func() {
		i.l.Debug("开启消费协程")
		err = cg.Consume(context.Background(), []string{InteractiveTopic}, saramax.NewHandler[InteractiveEvent](i.l, i.Consume))
		if err != nil {
			i.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()

	return nil
}

func (i *InteractiveConsumer) Consume(sg *sarama.ConsumerMessage, evt InteractiveEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100000*time.Second)
	defer cancel()
	i.l.Debug(fmt.Sprintf("开始消费 %v", evt))
	handleFunc, ok := i.handlers[evt.Type]
	if !ok {
		i.l.Error(fmt.Sprintf("未找到 handler for type %v", evt.Type))
		return nil
	}
	return handleFunc.Handle(ctx, evt)
}

type InteractiveEvent struct {
	Type  int64  `json:"type,omitempty"`
	Uid   int64  `json:"uid"`
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
}

type InteractiveHandler interface {
	Handle(ctx context.Context, data InteractiveEvent) error
}

type LikeHandler struct {
	syncSvc service.SyncService
}

var _ InteractiveHandler = (*LikeHandler)(nil)

func (l *LikeHandler) Handle(ctx context.Context, data InteractiveEvent) error {
	return handle(ctx, l.syncSvc, "like_index", getDocId(data), data)
}

type CollectHandler struct {
	syncSvc service.SyncService
}

var _ InteractiveHandler = (*CollectHandler)(nil)

func (c *CollectHandler) Handle(ctx context.Context, data InteractiveEvent) error {
	return handle(ctx, c.syncSvc, "collect_index", getDocId(data), data)
}

type CancelLikeHandler struct {
	syncSvc service.SyncService
}

var _ InteractiveHandler = (*CancelLikeHandler)(nil)

func (cl *CancelLikeHandler) Handle(ctx context.Context, data InteractiveEvent) error {
	return cl.syncSvc.Delete(ctx, "like_index", getDocId(data))
}

type CancelCollectHandler struct {
	syncSvc service.SyncService
}

var _ InteractiveHandler = (*CancelCollectHandler)(nil)

func (cc *CancelCollectHandler) Handle(ctx context.Context, data InteractiveEvent) error {
	return cc.syncSvc.Delete(ctx, "collect_index", getDocId(data))
}

func getDocId(data InteractiveEvent) string {
	return fmt.Sprintf("%d_%s_%d", data.Uid, data.Biz, data.BizId)
}

func handle(ctx context.Context, syncSvc service.SyncService, index, key string, data any) error {
	val, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return syncSvc.InputAny(ctx, index, key, string(val))
}
