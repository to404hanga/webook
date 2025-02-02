//go:build wireinject

package main

import (
	"webook/ranking/grpc"
	"webook/ranking/ioc"
	"webook/ranking/repository"
	"webook/ranking/repository/cache"
	"webook/ranking/service"

	"github.com/google/wire"
)

var serviceProviderSet = wire.NewSet(
	cache.NewRankingLocalCache,
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitInterActiveRpcClient,
	ioc.InitArticleRpcClient,
	ioc.InitEtcdClient,
	ioc.InitLogger,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		serviceProviderSet,
		grpc.NewRankingServiceServer,
		ioc.InitGrpcServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
