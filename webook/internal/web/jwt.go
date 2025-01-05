package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtHandler struct {
	signingMethod jwt.SigningMethod
	refreshKey    []byte
}

func NewJwtHandler() *jwtHandler {
	return &jwtHandler{
		signingMethod: jwt.SigningMethodHS512,
		refreshKey:    []byte("EZUAnsruwZew6sVuEX0hRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP8G"),
	}
}

var (
	JWTKey = []byte("EZUAnsruwZew6sVuEXUhRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP8G")
)

func ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}
	return segs[1]
}

func (h *jwtHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	uc := UserClaims{
		UserId:    uid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)), // 30 分钟过期
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *jwtHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	rc := RefreshClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 7 天过期
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(h.refreshKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid int64
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	UserAgent string
}
