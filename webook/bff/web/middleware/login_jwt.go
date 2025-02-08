package middleware

import (
	"net/http"
	myJwt "webook/bff/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	myJwt.Handler
}

func NewLoginJWTMiddlewareBuilder(handler myJwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: handler,
	}
}

var paths = []string{
	"/users/signup",
	"/users/login",
	"/users/login_sms/code/send",
	"/users/login_sms",
	"/users/refresh_token",
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
		tokenStr := m.ExtractToken(ctx)
		var uc myJwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(t *jwt.Token) (interface{}, error) {
			return myJwt.JWTKey, nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if err = m.CheckSession(ctx, uc.Ssid); err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)
	}
}
