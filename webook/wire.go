//go:build wireinject

package main

import (
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		dao.NewUserDAO,
		cache.NewCodeCache, cache.NewUserCache,
		repository.NewUserRepository, repository.NewCodeRepository,
		ioc.InitSMSService, service.NewUserService, service.NewCodeService,
		web.NewUserHandler,
		ioc.InitGinMiddlewares, ioc.InitWebServer,
	)
	return gin.Default()
}
