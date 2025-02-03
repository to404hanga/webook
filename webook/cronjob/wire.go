//go:build wireinject

package main

import (
	"webook/cronjob/grpc"
	"webook/cronjob/ioc"
	"webook/cronjob/repository"
	"webook/cronjob/repository/dao"
	"webook/cronjob/service"

	"github.com/google/wire"
)

var svcProviderSet = wire.NewSet(
	dao.NewGormJobDAO,
	repository.NewPreemptJobRepository,
	service.NewCronJobService,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitEtcdClient,
	ioc.InitLogger,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		svcProviderSet,
		grpc.NewCronJobServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
