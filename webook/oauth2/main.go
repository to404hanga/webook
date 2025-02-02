package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/grpcx"
)

func main() {
	initViperWatch()
	app := Init()
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
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
	server *grpcx.Server
}
