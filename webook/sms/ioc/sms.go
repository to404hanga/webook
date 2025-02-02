package ioc

import (
	"webook/sms/service"
	"webook/sms/service/localsms"
	"webook/sms/service/tencent"

	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/to404hanga/pkg404/logger"
)

func InitSmsTencentService(l logger.Logger) service.Service {
	type Config struct {
		SecretID  string `yaml:"secretId"`
		SecretKey string `yaml:"secretKey"`
	}
	var cfg Config
	err := viper.UnmarshalKey("tencentSms", &cfg)
	if err != nil {
		panic(err)
	}
	c, err := tencentSMS.NewClient(common.NewCredential(cfg.SecretID, cfg.SecretKey),
		"ap-nanjing",
		profile.NewClientProfile())
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, "1400842696", "妙影科技", l)
}

func InitSmsService() service.Service {
	return InitSmsMemoryService()
}

// InitSmsMemoryService 使用基于内存，输出到控制台的实现
func InitSmsMemoryService() service.Service {
	return localsms.NewService()
}
