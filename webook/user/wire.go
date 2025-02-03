//go:build wireinject

package main

import (
	"webook/pkg_local/wego"
	"webook/user/grpc"
	"webook/user/ioc"
	"webook/user/repository"
	"webook/user/repository/cache"
	"webook/user/repository/dao"
	"webook/user/service"

	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitEtcdClient,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		cache.NewRedisUserCache,
		dao.NewGormUserDAO,
		repository.NewCachedUserRepository,
		service.NewUserService,
		grpc.NewUserServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
