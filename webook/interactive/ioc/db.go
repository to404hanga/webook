package ioc

import (
	dao2 "webook/interactive/repository/dao"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	prometheus2 "github.com/to404hanga/pkg404/gormx/callbacks/prometheus"
	"github.com/to404hanga/pkg404/gormx/connpool"
	"github.com/to404hanga/pkg404/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	prometheusG "gorm.io/plugin/prometheus"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB() SrcDB {
	return initDB("src")
}

func InitDstDB() DstDB {
	return initDB("dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.Logger) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}

func InitBizDB(p *connpool.DoubleWritePool) *gorm.DB {
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: p,
	}))
	if err != nil {
		panic(err)
	}
	return doubleWrite
}

func initDB(key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		DSN: "root:root@tcp(localhost:3306)/webook",
	}
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.Use(prometheusG.New(prometheusG.Config{
		DBName:          "webook_" + key,
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

	cb := prometheus2.NewCallbacks(prometheus.SummaryOpts{
		Namespace: "to404hanga_lsh",
		Subsystem: "webook",
		Name:      "gorm_db_" + key,
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

	err = cb.Initialize(db)
	if err != nil {
		panic(err)
	}

	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithDBName("webook_"+key)))
	if err != nil {
		panic(err)
	}

	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}

	return db
}
