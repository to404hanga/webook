//go:build wireinject

package startup

import (
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	myJwt "webook/internal/web/jwt"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		InitRedis, ioc.InitDB,
		dao.NewUserDAO,

		cache.NewCodeCache, cache.NewUserCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		myJwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
