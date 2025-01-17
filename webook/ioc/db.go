package ioc

import (
	dao2 "webook/interactive/repository/dao"
	"webook/internal/repository/dao"
	"webook/pkg/gormx"
	"webook/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	prometheusG "gorm.io/plugin/prometheus"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		DSN: "root:root@tcp(localhost:3316)/webook",
	}
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(goormLoggerFunc(l.Debug), glogger.Config{
			// 慢查询
			SlowThreshold: 0,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}
	err = db.Use(prometheusG.New(prometheusG.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []prometheusG.MetricsCollector{
			&prometheusG.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	cb := gormx.NewCallbacks(prometheus.SummaryOpts{
		Namespace: "to404hanga_lsh",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计 gorm 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithDBName("webook")))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}

	// TODO 拆分为微服务后删除这块
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}

	return db
}

type goormLoggerFunc func(msg string, fields ...logger.Field)

func (g goormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Val: i})
}
