package ioc

import (
	"fmt"
	"webook/comment/repository/dao"

	"github.com/spf13/viper"
	prometheus2 "github.com/to404hanga/pkg404/gormx/callbacks/prometheus"
	"github.com/to404hanga/pkg404/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	c := Config{
		DSN: "root:123456@tcp(localhost:3307)/webook_follow",
	}
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(fmt.Errorf("初始化配置失败 %v, 原因 %w", c, err))
	}
	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		//使用 DEBUG 来打印
		Logger: glogger.Default.LogMode(glogger.Info),
	})
	if err != nil {
		panic(err)
	}

	// 接入 prometheus
	err = db.Use(prometheus.New(prometheus.Config{
		DBName: "webook",
		// 每 15 秒采集一些数据
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
