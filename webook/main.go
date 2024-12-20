package main

import (
	"strings"
	"time"
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()

	server := initWebServer()

	initUser(db, server)

	// server := gin.Default()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(200, "Hello, World!")
	})

	server.Run(":8080")
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Jwt-Token", "X-Refresh-Token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "127.0.0.1")
		},
		MaxAge: 12 * time.Hour,
	}))

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	// 限流: 每秒限流 100 个请求
	server.Use(ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 100)).Builder())

	useJWT(server)

	return server
}

func initUser(db *gorm.DB, server *gin.Engine) {
	userDAO := dao.NewUserDAO(db)
	userRepository := repository.NewUserRepository(userDAO)
	userService := service.NewUserService(userRepository)
	userHandler := web.NewUserHandler(userService)
	userHandler.RegisterRoutes(server)
}

func useJWT(server *gin.Engine) {
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

func useSession(server *gin.Engine) {
	store := cookie.NewStore([]byte("secret"))
	// 基于内存的实现
	// 传入两个密钥，第一个用于身份验证，第二个用于加密
	// store := memstore.NewStore([]byte("6zpKQvqguzUG92Hx4Thp9pE3KBkpoWdYpq0fvk05MaV6ehT0aZZBDFL9rxh8W5Qs"), []byte("oFZvqU3WsUiogDuvtLFNZaLVpGjQnwehzowpiQWk9gx9geikC6h6EtLK3sFctTau"))
	// store, err := redis.NewStore(16, "tcp", "localhost:16379", "",
	// 	[]byte("6zpKQvqguzUG92Hx4Thp9pE3KBkpoWdYpq0fvk05MaV6ehT0aZZBDFL9rxh8W5Qs"),
	// 	[]byte("oFZvqU3WsUiogDuvtLFNZaLVpGjQnwehzowpiQWk9gx9geikC6h6EtLK3sFctTau"),
	// )
	// if err != nil {
	// 	panic(err)
	// }

	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
}
