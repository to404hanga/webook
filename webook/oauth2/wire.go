//go:build wireinject

package main

import (
	"webook/oauth2/grpc"
	"webook/oauth2/ioc"

	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitEtcdClient,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		ioc.InitPrometheus,
		grpc.NewOauth2ServiceServer,
		ioc.InitGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
