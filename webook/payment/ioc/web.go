package ioc

import (
	"webook/payment/web"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/ginx"
)

func InitGinServer(handler *web.WechatHandler) *ginx.Server {
	engine := gin.Default()
	handler.RegisterRoutes(engine)
	addr := viper.GetString("http.addr")
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "to404hanga_lsh",
		Subsystem: "webook_payment",
		Name:      "http",
	})
	return &ginx.Server{
		Engine: engine,
		Addr:   addr,
	}
}
