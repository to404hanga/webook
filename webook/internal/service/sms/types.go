package sms

import "context"

// Service 发送短信的抽象接口
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
