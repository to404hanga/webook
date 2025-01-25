//go:build wireinject

package main

import (
	"webook/comment/grpc"
	"webook/comment/ioc"
	"webook/comment/repository"
	"webook/comment/repository/dao"
	"webook/comment/service"

	"github.com/google/wire"
)

var svcProviderSet = wire.NewSet(
	dao.NewCommentGormDAO,
	repository.NewCachedCommentRepository,
	service.NewCommentService,
	grpc.NewGrpcServer,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitEtcdClient,
	ioc.InitDB,
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
