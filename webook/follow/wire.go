//go:build wireinject

package main

import (
	"webook/follow/grpc"
	"webook/follow/ioc"
	"webook/follow/repository"
	"webook/follow/repository/cache"
	"webook/follow/repository/dao"
	"webook/follow/service"

	"github.com/google/wire"
)

var svcProviderSet = wire.NewSet(
	cache.NewFollowRedisCache,
	dao.NewFollowRelationGormDAO,
	repository.NewCachedFollowRepository,
	service.NewFollowRelationService,
	grpc.NewFollowRelationServiceServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitDB,
	ioc.InitRedis,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		svcProviderSet,
		ioc.InitGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
