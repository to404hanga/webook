package ioc

import (
	"context"
	"os"
	"webook/payment/events"
	"webook/payment/repository"
	"webook/payment/service/wechat"

	"github.com/to404hanga/pkg404/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WechatConfig struct {
	AppID        string
	MchID        string
	MchKey       string
	MchSerialNum string
	CertPath     string // 证书
	KeyPath      string
}

func InitWechatClient(cfg WechatConfig) *core.Client {
	// 使用 utils 提供的函数从本地文件中加载商户私钥，用于生成请求的签名
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(cfg.KeyPath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client, err := core.NewClient(ctx, option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.MchSerialNum, mchPrivateKey, cfg.MchKey))
	if err != nil {
		panic(err)
	}
	return client
}

func InitWechatNativeService(cli *core.Client, repo repository.PaymentRepository, producer events.Producer, l logger.Logger, cfg WechatConfig) *wechat.NativePaymentService {
	return wechat.NewNativePaymentService(cfg.AppID, cfg.MchID, producer, repo, &native.NativeApiService{
		Client: cli,
	}, l)
}

func InitWechatNotifyHandler(cfg WechatConfig) *notify.Handler {
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
	handler, err := notify.NewRSANotifyHandler(cfg.MchKey, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	if err != nil {
		panic(err)
	}
	return handler
}

func InitWechatConfig() WechatConfig {
	return WechatConfig{
		AppID:        os.Getenv("WEPAY_APP_ID"),
		MchID:        os.Getenv("WEPAY_MCH_ID"),
		MchKey:       os.Getenv("WEPAY_MCH_KEY"),
		MchSerialNum: os.Getenv("WEPAY_MCH_SERIAL_NUM"),
		CertPath:     "./config/cert/apiclient_cert.pem",
		KeyPath:      "./config/cert/apiclient_key.pem",
	}
}
