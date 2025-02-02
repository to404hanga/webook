//go:build wireinject

package main

import (
	"webook/pkg_local/wego"
	"webook/sms/grpc"
	"webook/sms/ioc"

	"github.com/google/wire"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitSmsMemoryService,
		grpc.NewSmsServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
