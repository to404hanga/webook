//go:build wireinject

package main

import (
	"webook/code/grpc"
	"webook/code/ioc"
	"webook/code/repository"
	"webook/code/repository/cache"
	"webook/code/service"

	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitRedis,
	ioc.InitLogger,
	ioc.InitEtcdClient,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		ioc.InitSmsRpcClient,
		cache.NewRedisCodeCache,
		repository.NewCachedCodeRepository,
		service.NewSMSCodeService,
		grpc.NewCodeServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
