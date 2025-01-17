package ratelimit

import (
	"context"
	"errors"
	"testing"
	"webook/internal/service/sms"
	smsmocks "webook/internal/service/sms/mocks"
	"webook/pkg/limiter"
	limitermocks "webook/pkg/limiter/mocks"

	"github.com/tj/assert"
	"go.uber.org/mock/gomock"
)

func TestRateLimiterSMSService(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter)
		wantErr error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitermocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return svc, limiter
			},
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitermocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, limiter
			},
			wantErr: ErrLimited,
		},
		{
			name: "限流器错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				limiter := limitermocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("redis限流器错误"))
				return svc, limiter
			},
			wantErr: errors.New("redis限流器错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			smsSvc, limiter := tc.mock(ctrl)
			svc := NewRateLimiterSMSService(smsSvc, limiter)
			err := svc.Send(context.Background(), "abc", []string{"123"}, "123456")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
