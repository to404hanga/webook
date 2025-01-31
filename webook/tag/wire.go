//go:build wireinject

package main

import (
	"webook/pkg_local/wego"
	"webook/tag/grpc"
	"webook/tag/ioc"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"
	"webook/tag/service"

	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitEtcdClient,
	ioc.InitKafka,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		cache.NewRedisTagCache,
		dao.NewGormTagDAO,
		ioc.InitRepository,
		ioc.InitProducer,
		service.NewTagService,
		grpc.NewTagGrpcServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
