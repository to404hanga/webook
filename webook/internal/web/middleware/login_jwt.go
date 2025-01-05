package middleware

import (
	"net/http"
	"webook/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
}

var paths = []string{
	"/users/signup",
	"/users/login",
	"/users/login_sms/code/send",
	"/users/login_sms",
	"/oauth2/wechat/authurl",
	"/oauth2/wechat/callback",
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		for _, p := range paths {
			if p == path {
				return
			}
		}
		tokenStr := web.ExtractToken(ctx)
		var uc web.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(t *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("user", uc)
	}
}
