package sms

import "context"

//go:generate mockgen -source=./types.go -package=smsmocks -destination=./mocks/types.mock.go Service
// Service 发送短信的抽象接口
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
