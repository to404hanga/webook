package events

import (
	"context"
	"strconv"
	"time"
	"webook/im/domain"
	"webook/im/service"

	"github.com/IBM/sarama"
	"github.com/to404hanga/pkg404/canalx"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/saramax"
)

type MySQLBinlogConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.UserService
}

func (m *MySQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("im", m.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(), []string{"webook_binlog"}, saramax.NewHandler[canalx.Message[User]](m.l, m.Consume))
		if err != nil {
			m.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (m *MySQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage, val canalx.Message[User]) error {
	if val.Table != "users" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range val.Data {
		err := m.svc.Sync(ctx, domain.User{
			Nickname: data.Nickname,
			UserID:   strconv.FormatInt(data.Id, 10),
		})
		if err != nil {
			m.l.Error("同步用户信息失败", logger.Int64("id", data.Id), logger.Error(err))
			continue
		}
	}
	return nil
}

type User struct {
	Id            int64  `json:"id"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Nickname      string `json:"nickname"`
	Birthday      int64  `json:"birthday"`
	AboutMe       string `json:"about_me"`
	Phone         string `json:"phone"`
	WechatOpenId  string `json:"wechat_open_id"`
	WechatUnionId string `json:"wechat_union_id"`
	CreateTime    int64  `json:"create_time"`
	UpdateTime    int64  `json:"update_time"`
}
