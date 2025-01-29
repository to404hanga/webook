package main

import (
	"webook/search/events"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/grpcx"
)

func main() {
	initViperWatch()
	app := Init()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	panic(err)
}

func initViperWatch() {
	cfile := pflag.String("config",
		"config/dev.yaml", "配置文件路径")
	pflag.Parse()
	// 直接指定文件路径
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

type App struct {
	server    *grpcx.Server
	consumers []events.Consumer
}
