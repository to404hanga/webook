package ioc

import (
	"os"
	"webook/internal/service/oauth2/wechat"

	"github.com/to404hanga/pkg404/logger"
)

func InitWechatService(l logger.Logger) wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appID, appSecret, l)
}
