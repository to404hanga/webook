//go:build wireinject

package main

import (
	"webook/feed/events"
	"webook/feed/grpc"
	"webook/feed/ioc"
	"webook/feed/repository"
	"webook/feed/repository/cache"
	"webook/feed/repository/dao"
	"webook/feed/service"

	"github.com/google/wire"
)

var svcProviderSet = wire.NewSet(
	dao.NewFeedPushEventDAO,
	dao.NewFeedPullEventDAO,
	cache.NewFeedEventRedisCache,
	repository.NewCachedFeedEventRepository,
)

var thirdProvider = wire.NewSet(
	ioc.InitEtcdClient,
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitKafka,
	ioc.InitDB,
	ioc.InitFollowClient,
)

func Init() *App {
	wire.Build(
		thirdProvider,
		svcProviderSet,
		ioc.RegisterHandler,
		service.NewFeedService,
		grpc.NewFeedServiceServer,
		events.NewFeedEventConsumer,
		events.NewArticleEventConsumer,
		ioc.InitGrpcxServer,
		ioc.NewConsumers,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
