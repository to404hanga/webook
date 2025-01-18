package ioc

import (
	"webook/interactive/repository/dao"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/ginx"
	"github.com/to404hanga/pkg404/gormx/connpool"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/migrator/events"
	"github.com/to404hanga/pkg404/migrator/events/fixer"
	"github.com/to404hanga/pkg404/migrator/scheduler"
)

func InitGinxServer(l logger.Logger, src SrcDB, dst DstDB, pool *connpool.DoubleWritePool, producer events.Producer) *ginx.Server {
	engine := gin.Default()
	group := engine.Group("/migrator")
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "to404hanga_lsh",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	sch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer("inconsistent_interactive", p)
}

func InitFixerConsumer(client sarama.Client, l logger.Logger, src SrcDB, dst DstDB) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
