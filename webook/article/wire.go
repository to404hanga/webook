//go:build wireinject

package main

import (
	"webook/article/events"
	"webook/article/grpc"
	"webook/article/ioc"
	"webook/article/repository"
	"webook/article/repository/cache"
	"webook/article/repository/dao"
	"webook/article/service"
	"webook/pkg_local/wego"

	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitUserRpcClient,
	ioc.InitProducer,
	ioc.InitEtcdClient,
	ioc.InitDB,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		events.NewSaramaSyncProducer,
		cache.NewArticleRedisCache,
		dao.NewGormArticleDAO,
		repository.NewArticleRepository,
		repository.NewGrpcAuthorRepository,
		service.NewArticleService,
		grpc.NewArticleServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
