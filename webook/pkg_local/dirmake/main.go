package main

import (
	"fmt"
	"os"
	"strings"
)

var dirs = []string{
	"/config",
	"/domain",
	"/ioc",
	"/service",
}

func main() {
	args := os.Args
	if len(args) <= 1 {
		panic("应输入至少一个参数")
	}

	prefixDir := ""
	grpc := false
	web := false
	events := false
	test := false
	dao := false
	cache := false
	for _, arg := range args {
		if prefixDir == "" && strings.HasPrefix(arg, "--source=") {
			prefixDir = strings.TrimPrefix(arg, "--source=")
		}
		if !grpc {
			if strings.HasPrefix(arg, "--grpc=") {
				tmp := strings.TrimPrefix(arg, "--grpc=")
				switch tmp {
				case "true", "True", "T", "1", "TRUE":
					grpc = true
				}
			} else if arg == "-g" {
				grpc = true
			}
		}
		if !web {
			if strings.HasPrefix(arg, "--web=") {
				tmp := strings.TrimPrefix(arg, "--web=")
				switch tmp {
				case "true", "True", "T", "1", "TRUE":
					web = true
				}
			} else if arg == "-w" {
				web = true
			}
		}
		if !events {
			if strings.HasPrefix(arg, "--events=") {
				tmp := strings.TrimPrefix(arg, "--events=")
				switch tmp {
				case "true", "True", "T", "1", "TRUE":
					events = true
				}
			} else if arg == "-e" {
				events = true
			}
		}
		if !test {
			if strings.HasPrefix(arg, "--test=") {
				tmp := strings.TrimPrefix(arg, "--test=")
				switch tmp {
				case "true", "True", "T", "1", "TRUE":
					test = true
				}
			} else if arg == "-t" {
				test = true
			}
		}
		if !dao {
			if strings.HasPrefix(arg, "--dao=") {
				tmp := strings.TrimPrefix(arg, "--dao=")
				switch tmp {
				case "true", "True", "T", "1", "TRUE":
					dao = true
				}
			} else if arg == "-d" {
				dao = true
			}
		}
		if !cache {
			if strings.HasPrefix(arg, "--cache=") {
				tmp := strings.TrimPrefix(arg, "--cache=")
				switch tmp {
				case "true", "True", "T", "1", "TRUE":
					cache = true
				}
			} else if arg == "-c" {
				cache = true
			}
		}
	}

	if prefixDir == "" {
		panic("未找到 --source=参数")
	}

	tmp := strings.Split(prefixDir, "/")
	filename := tmp[len(tmp)-1]

	filenames := []string{}

	if grpc {
		dirs = append(dirs, "/grpc")
		filenames = append(filenames, "/grpc/"+filename+".go")
	}
	if web {
		dirs = append(dirs, "/web")
		filenames = append(filenames, "/web/"+filename+".go")
	}
	if events {
		dirs = append(dirs, "/events")
	}
	if test {
		dirs = append(dirs, "/integration/startup")
	}
	if dao {
		dirs = append(dirs, "/repository/dao")
		filenames = append(filenames, "/repository/dao/"+filename+".go")
		filenames = append(filenames, "/repository/dao/init.go")
		filenames = append(filenames, "/repository/"+filename+".go")
	}
	if cache {
		dirs = append(dirs, "/repository/cache")
		filenames = append(filenames, "/repository/cache/"+filename+".go")
		if !dao {
			filenames = append(filenames, "/repository/"+filename+".go")
		}
	}

	filenames = append(filenames, "/service/"+filename+".go")
	filenames = append(filenames, "/domain/"+filename+".go")

	for _, dir := range dirs {
		fmt.Println(prefixDir + dir)
		os.MkdirAll(prefixDir+dir, 0777)
	}

	for _, file := range filenames {
		fmt.Println(prefixDir + file)
		tmp := strings.Split(file, "/")
		packageName := tmp[len(tmp)-2]
		os.Create(prefixDir + file)
		f, err := os.OpenFile(prefixDir+file, os.O_RDWR|os.O_TRUNC, 0777)
		if err != nil {
			panic(err)
		}
		f.WriteString("package " + packageName + "\n")
		f.Close()
	}

	os.Create(prefixDir + "/main.go")
	os.Create(prefixDir + "/wire.go")
	os.Create(prefixDir + "/config/dev.yaml")

	file, err := os.OpenFile(prefixDir+"/main.go", os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}
	file.WriteString(`package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViperWatch()
	app := Init()
	err := app.GRPCServer.Serve()
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
}`)
	file.Close()

	file, err = os.OpenFile(prefixDir+"/wire.go", os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}
	file.WriteString(`//go:build wireinject

package main
`)
	file.Close()

	if dao {
		f, err := os.OpenFile(prefixDir+"/repository/dao/init.go", os.O_RDWR|os.O_APPEND, 0777)
		if err != nil {
			panic(err)
		}
		f.WriteString(`
import "gorm.io/gorm"

func InitTables(db *gorm.DB) error{
	return db.AutoMigrate(
		
	)
}`)
		f.Close()
		f, err = os.OpenFile(prefixDir+"/ioc/db.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			panic(err)
		}
		f.WriteString(`package ioc

import (
	"fmt"
	"webook/` + filename + `/repository/dao"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string ` + "`" + `yaml:"dsn"` + "`" + `
	}
	c := Config{
		DSN: "root:123456@tcp(localhost:3306)/webook_sms",
	}
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(fmt.Errorf("初始化配置失败 %v，原因 %v", c, err))
	}
	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
`)
		f.Close()
	}
	if cache {
		f, err := os.OpenFile(prefixDir+"/ioc/redis.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			panic(err)
		}
		f.WriteString(`package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	cmd := redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis.addr"),
	})
	return cmd
}
`)
		f.Close()
	}
	if grpc {
		f, err := os.OpenFile(prefixDir+"/ioc/etcd.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			panic(err)
		}
		f.WriteString(`package ioc

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcdClient() *clientv3.Client {
	var cfg clientv3.Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		panic(err)
	}
	return client
}
`)
		f.Close()
	}
	file, err = os.OpenFile(prefixDir+"/ioc/logger.go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		panic(err)
	}
	file.WriteString(`package ioc

import (
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/logger"
	"go.uber.org/zap"
)

func InitLogger() logger.Logger {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}`)
}
