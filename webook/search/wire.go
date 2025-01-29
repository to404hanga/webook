//go:build wireinject

package main

import (
	"webook/search/events"
	"webook/search/grpc"
	"webook/search/ioc"
	"webook/search/repository"
	"webook/search/repository/dao"
	"webook/search/service"

	"github.com/google/wire"
)

var svcProviderSet = wire.NewSet(
	dao.NewUserElasticSearchDAO,
	dao.NewArticleElasticSearchDAO,
	dao.NewAnyElasticSearchDAO,
	dao.NewTagElasticSearchDAO,
	repository.NewUserRepository,
	repository.NewArticleRepository,
	repository.NewAnyRepository,
	service.NewSyncService,
	service.NewSearchService,
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitESClient,
	ioc.InitEtcdClient,
	ioc.InitKafka,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		svcProviderSet,
		grpc.NewSyncServiceServer,
		grpc.NewSearchServiceServer,
		events.NewUserConsumer,
		events.NewArticleConsumer,
		ioc.InitGrpcxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
