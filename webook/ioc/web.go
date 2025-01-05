package ioc

import (
	"strings"
	"time"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitWebServer(middlewares []gin.HandlerFunc, userHandler *web.UserHandler, wechatHandler *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	userHandler.RegisterRoutes(server)
	wechatHandler.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			//	AllowCredentials: true,
			//	AllowHeaders:     []string{"Content-Type", "Authorization"},
			//	ExposeHeaders:    []string{"X-Jwt-Token", "X-Refresh-Token"},
			AllowOriginFunc: func(origin string) bool {
				if strings.HasPrefix(origin, "http://localhost") {
					return true
				}
				return strings.Contains(origin, "192.168.31.102")
			},
			MaxAge: 12 * time.Hour,
		}),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Builder(),
		(&middleware.LoginJWTMiddlewareBuilder{}).CheckLogin(),
	}
}
