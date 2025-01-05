package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtHandler struct {
}

var (
	JWTKey = []byte("EZUAnsruwZew6sVuEXUhRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP8G")
)

func (h *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) {
	uc := UserClaims{
		UserId:    uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 30 分钟过期
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header("x-jwt-token", tokenStr)
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	UserAgent string
}
