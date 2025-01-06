package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"webook/internal/integration/startup"
	"webook/internal/web"
	"webook/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/tj/assert"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestUserHandler_SendSMSCode(t *testing.T) {
	rdb := startup.InitRedis()
	server := startup.InitWebServer()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		phone    string
		wantCode int
		wantBody ginx.Result
	}{
		{
			name:   "发送成功的用例",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				key := "phone_code:login:1234567890"
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) > 0)
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > 9*time.Minute+50*time.Second)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    "1234567890",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 200,
				Msg:  "发送成功",
			},
		},
		{
			name:     "未输入手机号码",
			before:   func(t *testing.T) {},
			after:    func(t *testing.T) {},
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 400,
				Msg:  "请输入手机号",
			},
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				key := "phone_code:login:1234567890"
				err := rdb.Set(ctx, key, "123456", 9*time.Minute+50*time.Second).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				key := "phone_code:login:1234567890"
				// code, err := rdb.GetDel(ctx, key).Result()
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    "1234567890",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 429,
				Msg:  "短信发送太频繁，请稍后再试",
			},
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				key := "phone_code:login:1234567890"
				err := rdb.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				key := "phone_code:login:1234567890"
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    "1234567890",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Code: 500,
				Msg:  "系统错误",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			req, err := http.NewRequest(http.MethodPost, "/users/login_sms/code/send", bytes.NewReader([]byte(fmt.Sprintf(`{"phone": "%s"}`, tc.phone))))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res web.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}
