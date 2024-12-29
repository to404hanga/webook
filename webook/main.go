package main

import (
	"webook/internal/web/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {

	server := InitWebServer()

	server.Run(":8080")
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
