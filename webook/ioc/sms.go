package ioc

import (
	"os"
	"webook/internal/service/sms"
	localsms "webook/internal/service/sms/localsms"
	"webook/internal/service/sms/tencent"
	"webook/pkg/logger"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

func InitSMSService() sms.Service {
	//return ratelimit.NewRateLimitSMSService(localsms.NewService(), limiter.NewRedisSlidingWindowLimiter())
	return localsms.NewService()
	// 如果有需要，就可以用这个
	//return initTencentSMSService()
}

func initTencentSMSService(l logger.Logger) sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯 SMS 的 secret key")
	}
	c, err := tencentSMS.NewClient(
		common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile(),
	)
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, "1400842696", "妙影科技", l)
}
