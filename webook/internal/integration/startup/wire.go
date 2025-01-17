//go:build wireinject

package startup

import (
	repository2 "webook/interactive/repository"
	cache2 "webook/interactive/repository/cache"
	dao2 "webook/interactive/repository/dao"
	service2 "webook/interactive/service"
	"webook/internal/events/article"
	"webook/internal/job"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	articleDao "webook/internal/repository/dao/article"
	"webook/internal/service"
	"webook/internal/service/sms"
	"webook/internal/service/sms/async"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitSaramaClient,
	InitSyncProducer,
	InitLogger,
)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGormJobDAO)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	repository.NewCachedArticleRepository,
	cache.NewArticleRedisCache,
	articleDao.NewGormArticleDAO,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGormInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcSet,
		// cache 部分
		cache.NewCodeCache,

		// repository 部分
		repository.NewCachedCodeRepository,

		article.NewSaramaSyncProducer,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitAsyncSmsService(svc sms.Service) *async.Service {
	wire.Build(
		thirdPartySet,
		repository.NewAsyncSmsRepository,
		dao.NewGormAsyncSmsDAO,
		async.NewService,
	)
	return &async.Service{}
}

func InitArticleHandler(dao articleDao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		interactiveSvcSet,
		repository.NewCachedArticleRepository,
		cache.NewArticleRedisCache,
		service.NewArticleService,
		article.NewSaramaSyncProducer,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service2.NewInteractiveService(nil)
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdPartySet, job.NewScheduler)
	return &job.Scheduler{}
}
