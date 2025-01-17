package startup

import (
	"webook/internal/service/oauth2/wechat"

	"github.com/to404hanga/pkg404/logger"
)

func InitWechatService(l logger.Logger) wechat.Service {
	return wechat.NewService("", "", l)
}
