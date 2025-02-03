package cache

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/tj/assert"
)

func TestCodeCache(t *testing.T) {
	// keyFunc := func(biz, phone string) string {
	// 	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	// }
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			// name: "设置成功",
			// mock: func(ctrl *gomock.Controller) redis.Cmdable {
			// 	res := redismocks.NewMockCmdable(ctrl)
			// 	cmd := redis.NewCmd(context.Background())
			// 	cmd.SetErr(nil)
			// 	cmd.SetVal(int64(0))
			// 	res.EXPECT().Eval(gomock.Any(), luaSetCode, []string{keyFunc("test", "1234567890")}, []interface{}{"123456"}).Return()
			// 	return res
			// },
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := NewRedisCodeCache(tc.mock(ctrl), nil)
			err := cc.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
