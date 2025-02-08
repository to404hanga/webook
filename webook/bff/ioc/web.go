package ioc

import (
	"context"
	"strings"
	"time"
	"webook/bff/web"
	ijwt "webook/bff/web/jwt"
	"webook/bff/web/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/ginx"
	"github.com/to404hanga/pkg404/logger"
)

func InitGinServer(l logger.Logger, jwtHandler ijwt.Handler, user *web.UserHandler, article *web.ArticleHandler, reward *web.RewardHandler) *ginx.Server {
	engine := gin.Default()
	engine.Use(
		corsHandler(),
		timeout(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHandler).CheckLogin(),
	)
	user.RegisterRoutes(engine)
	article.RegisterRoutes(engine)
	reward.RegisterRoutes(engine)
	addr := viper.GetString("http.addr")
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "to404hanga_lsh",
		Subsystem: "webook_bff",
		Name:      "http",
	})
	ginx.L = l
	return &ginx.Server{
		Engine: engine,
		Addr:   addr,
	}
}

func timeout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, ok := ctx.Request.Context().Deadline()
		if !ok {
			// 强制给超时时间
			newCtx, cancel := context.WithTimeout(ctx.Request.Context(), time.Second*10)
			defer cancel()
			ctx.Request = ctx.Request.Clone(newCtx)
		}
		ctx.Next()
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
				return true
			}
			return strings.Contains(origin, "xxx.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
