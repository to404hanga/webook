//go:build wireinject

package main

import (
	"webook/interactive/events"
	"webook/interactive/grpc"
	"webook/interactive/ioc"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"

	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitSaramaClient,
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGormInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
