//go:build wireinject

package main

import (
	"webook/bff/ioc"
	"webook/bff/web"
	"webook/bff/web/jwt"
	"webook/pkg_local/wego"

	"github.com/google/wire"
)

func InitApp() *wego.App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitRedis,
		ioc.InitEtcdClient,

		web.NewArticleHandler,
		web.NewUserHandler,
		web.NewRewardHandler,
		jwt.NewRedisJWTHandler,

		ioc.InitUserClient,
		ioc.InitIntrClient,
		ioc.InitRewardClient,
		ioc.InitCodeClient,
		ioc.InitArticleClient,
		ioc.InitGinServer,
		wire.Struct(new(wego.App), "WebServer"),
	)
	return new(wego.App)
}
