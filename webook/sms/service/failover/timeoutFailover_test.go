package failover

import (
	"context"
	"errors"
	"testing"
	"webook/sms/service"
	svcmocks "webook/sms/service/mocks"

	"github.com/tj/assert"
	"go.uber.org/mock/gomock"
)

func TestTimeoutFailover_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) []service.Service
		threshold int64
		idx       int64
		cnt       int64
		wantErr   error
		wantCnt   int64
		wantIdx   int64
	}{
		{
			name: "没有触发切换",
			mock: func(ctrl *gomock.Controller) []service.Service {
				svc0 := svcmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []service.Service{svc0}
			},
			idx:       0,
			cnt:       12,
			threshold: 15,
			wantIdx:   0,
			wantCnt:   0,
			wantErr:   nil,
		},
		{
			name: "触发切换，成功",
			mock: func(ctrl *gomock.Controller) []service.Service {
				svc0 := svcmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				svc1 := svcmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return []service.Service{svc0, svc1}
			},
			idx:       0,
			cnt:       15,
			threshold: 15,
			wantIdx:   1,
			wantCnt:   0,
			wantErr:   nil,
		},
		{
			name: "触发切换，失败",
			mock: func(ctrl *gomock.Controller) []service.Service {
				svc0 := svcmocks.NewMockService(ctrl)
				svc1 := svcmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败")).AnyTimes()
				return []service.Service{svc0, svc1}
			},
			idx:       1,
			cnt:       15,
			threshold: 15,
			wantIdx:   0,
			wantCnt:   0,
			wantErr:   errors.New("发送失败"),
		},
		{
			name: "触发切换，超时",
			mock: func(ctrl *gomock.Controller) []service.Service {
				svc0 := svcmocks.NewMockService(ctrl)
				svc1 := svcmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).AnyTimes()
				return []service.Service{svc0, svc1}
			},
			idx:       1,
			cnt:       15,
			threshold: 15,
			wantIdx:   0,
			wantCnt:   1,
			wantErr:   context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewTimeoutFailoverSMSService(tc.mock(ctrl), tc.threshold)
			svc.cnt = tc.cnt
			svc.idx = tc.idx
			err := svc.Send(context.Background(), "1234", []string{"12", "34"}, "1234567890")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantCnt, svc.cnt)
			assert.Equal(t, tc.wantIdx, svc.idx)
		})
	}
}
