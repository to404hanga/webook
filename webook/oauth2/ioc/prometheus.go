package ioc

import (
	"webook/oauth2/service"

	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/logger"
)

func InitPrometheus(log logger.Logger) service.Service {
	svc := InitService(log)
	type Config struct {
		NameSpace  string `yaml:"nameSpace"`
		Subsystem  string `yaml:"subsystem"`
		InstanceID string `yaml:"instanceId"`
		Name       string `yaml:"name"`
	}
	var cfg Config
	err := viper.UnmarshalKey("prometheus", &cfg)
	if err != nil {
		panic(err)
	}
	return service.NewDecorator(svc, cfg.NameSpace, cfg.Subsystem, cfg.InstanceID, cfg.Name)
}
