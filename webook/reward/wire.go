//go:build wireinject

package main

import (
	"webook/pkg_local/wego"
	"webook/reward/grpc"
	"webook/reward/ioc"
	"webook/reward/repository"
	"webook/reward/repository/cache"
	"webook/reward/repository/dao"
	"webook/reward/service"

	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitRedis,
)

func Init() *wego.App {
	wire.Build(
		thirdPartySet,
		service.NewWechatNativeRewardService,
		ioc.InitAccountClient,
		ioc.InitGrpcxServer,
		ioc.InitPaymentClient,
		repository.NewRewardRepository,
		cache.NewRewardRedisCache,
		dao.NewRewardGormDAO,
		grpc.NewRewardServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
