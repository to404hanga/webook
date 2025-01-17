package tencent

import (
	"context"
	"fmt"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/to404hanga/pkg404/logger"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
	logger   logger.Logger
}

func NewService(client *sms.Client, appId, signName string, logger logger.Logger) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &signName,
		logger:   logger,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SetContext(ctx)
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.TemplateParamSet = s.toPtrSlice(args)
	req.PhoneNumberSet = s.toPtrSlice(numbers)

	res, err := s.client.SendSms(req)
	s.logger.Debug("请求腾讯SendSmd接口", logger.Any("req", req), logger.Any("resp", res))
	if err != nil {
		return err
	}

	for _, statusPtr := range res.Response.SendStatusSet {
		if statusPtr == nil {
			// 不可能进入这里
			continue
		}
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			// 发送失败
			return fmt.Errorf("发送短信失败 code: %s, msg: %s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data, func(idx int, src string) *string {
		return &src
	})
}
