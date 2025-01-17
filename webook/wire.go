//go:build wireinject

package main

import (
	"webook/interactive/events"
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	articleDao "webook/internal/repository/dao/article"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"

	"github.com/google/wire"
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGormInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRLockClient,
		// DAO 部分
		dao.NewUserDAO,
		articleDao.NewGormArticleDAO,

		interactiveSvcSet,
		ioc.InitIntrClient,
		rankingSvcSet,
		ioc.InitJobs,
		ioc.InitRankingJob,

		article.NewSaramaSyncProducer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleRedisCache,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCachedCodeRepository,
		repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
