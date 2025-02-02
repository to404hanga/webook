package ioc

import (
	"fmt"
	"webook/article/repository/dao"

	"github.com/spf13/viper"
	prometheus2 "github.com/to404hanga/pkg404/gormx/callbacks/prometheus"
	"github.com/to404hanga/pkg404/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	c := Config{
		DSN: "rot:123456@tcp(localhost:3306)/webook_article",
	}
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(fmt.Errorf("初始化数据库配置失败 %v，原因 %v", c, err))
	}

	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics()))
	if err != nil {
		panic(err)
	}

	prom := prometheus2.Callbacks{
		Namespace:  "to404hanga_lsh",
		Subsystem:  "webook",
		Name:       "gorm",
		InstanceId: "my-instance-1",
		Help:       "gorm DB 查询",
	}
	err = prom.Initialize(db)
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
