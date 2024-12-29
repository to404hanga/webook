package ioc

import (
	"webook/internal/service/sms"
	"webook/internal/service/sms/localSms"
)

func InitSMSService() sms.Service {
	return localSms.NewService()
}

// func InitTencentSMSService() sms.Service {
// 	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
// 	if !ok {
// 		panic("SMS_SECRET_ID not found")
// 	}
// 	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
// 	if !ok {
// 		panic("SMS_SECRET_KEY not found")
// 	}
// 	c, err := tencentSMS.NewClient(common.NewCredential(secretId, secretKey), "ap-nanjing", profile.NewClientProfile())
// 	if err != nil {
// 		panic(err)
// 	}
// 	return tencent.NewService(c, "1400491234", "腾讯云")
// }
