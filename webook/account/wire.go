//go:build wireinject

package main

import (
	"webook/account/grpc"
	"webook/account/ioc"
	"webook/account/repository"
	"webook/account/repository/dao"
	"webook/account/service"
	"webook/pkg_local/wego"

	"github.com/google/wire"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitGrpcxServer,
		dao.NewCreditGormDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
