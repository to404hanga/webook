package ioc

import (
	"webook/oauth2/service"

	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/logger"
)

func InitService(log logger.Logger) service.Service {
	type Config struct {
		AppID     string `yaml:"appId"`
		AppSecret string `yaml:"appSecret"`
	}
	var cfg Config
	err := viper.UnmarshalKey("weChatConf", &cfg)
	if err != nil {
		panic(err)
	}
	return service.NewService(cfg.AppID, cfg.AppSecret, log)
}
