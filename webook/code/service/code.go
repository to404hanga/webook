package service

import (
	"context"
	"fmt"
	"math/rand"
	smsv1 "webook/api/proto/gen/sms/v1"
	"webook/code/repository"
)

var (
	ErrCodeSendTooMany   = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany
)

const codeTplId = "1877556"

//go:generate mockgen -source=./code.go -package=svcmocks -destination=./mocks/code.mock.go CodeService
type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type SMSCodeService struct {
	repo repository.CodeRepository
	sms  smsv1.SmsServiceClient
}

var _ CodeService = (*SMSCodeService)(nil)

func NewSMSCodeService(repo repository.CodeRepository, sms smsv1.SmsServiceClient) CodeService {
	return &SMSCodeService{
		repo: repo,
		sms:  sms,
	}
}

func (svc *SMSCodeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	_, err = svc.sms.Send(ctx, &smsv1.SmsSendRequest{
		TplId:   codeTplId,
		Args:    []string{code},
		Numbers: []string{phone},
	})
	return err
}

func (svc *SMSCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		return false, nil
	}
	return ok, err
}

func (svc *SMSCodeService) generate() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
