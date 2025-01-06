package main

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	// initViperRemote()
	initViperWatch()
	initLogger()

	server := InitWebServer()

	server.Run(":8080")
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initViperWatch() {
	cfile := pflag.String("config", "./config/dev.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(viper.GetString("test.key"))
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	cfile := pflag.String("config", "./config/dev.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://localhost:2379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("json")

	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("远程配置中心发生变更")
	})

	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			log.Println("远程配置已更新")
			time.Sleep(time.Second)
		}
	}()
}
