package ratelimit

import (
	"context"
	"errors"
	"webook/internal/service/sms"
	"webook/pkg/limiter"
)

var (
	ErrLimited             = errors.New("触发限流")
	_          sms.Service = &RateLimiterSMSService{}
)

type RateLimiterSMSService struct {
	svc     sms.Service
	limiter limiter.Limiter
	key     string
}

func (r *RateLimiterSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	limited, err := r.limiter.Limit(ctx, r.key)
	if err != nil {
		return err
	}
	if limited {
		return ErrLimited
	}
	return r.svc.Send(ctx, tplId, args, number...)
}

func NewRateLimiterSMSService(svc sms.Service, limiter limiter.Limiter) *RateLimiterSMSService {
	return &RateLimiterSMSService{
		svc:     svc,
		limiter: limiter,
		key:     "sms-limiter",
	}
}
