package job

import (
	"context"
	"encoding/json"
	"time"
	"webook/payment/events"
	"webook/payment/repository"

	"github.com/to404hanga/pkg404/logger"
)

type LocalMsgResendJob struct {
	repo      repository.LocalMsgRepository
	producer  events.Producer
	threshold time.Duration
	l         logger.Logger
}

func NewLocalMsgResendJob(repo repository.LocalMsgRepository, producer events.Producer, threshold time.Duration, l logger.Logger) *LocalMsgResendJob {
	return &LocalMsgResendJob{
		repo:      repo,
		producer:  producer,
		threshold: threshold,
		l:         l,
	}
}

func (l *LocalMsgResendJob) Name() string {
	return "LocalMsgResendJob"
}

func (l *LocalMsgResendJob) Run() error {
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		msgs, err := l.repo.FindInitMsg(ctx, limit, offset)
		cancel()
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			var evt events.PaymentEvent
			err = json.Unmarshal([]byte(msg.Content), &evt)
			if err != nil {
				continue
			}
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
			err = l.producer.ProducePaymentEvent(ctx, evt)
			if err != nil {
				// 判断是否值得重试
				if msg.CreateTime.Add(l.threshold).Before(time.Now()) {
					err = l.repo.MarkFailed(ctx, msg.Id)
					if err != nil {
						l.l.Error("标记本地消息表为失败失败", logger.Error(err), logger.Int64("msg_id", msg.Id))
					}
				}
			} else {
				err = l.repo.MarkSuccess(ctx, msg.Id)
				if err != nil {
					l.l.Error("标记本地消息表为成功失败", logger.Error(err), logger.Int64("msg_id", msg.Id))
				}
			}
			cancel()
		}
		if len(msgs) < limit {
			return nil
		}
	}
}
