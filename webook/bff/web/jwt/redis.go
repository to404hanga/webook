package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

var (
	JWTKey           = []byte("EZUAnsruwZew6sVuEXUhRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP8G")
	RCJWTKey         = []byte("EZUAnsruwZew7sVuEXUhRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP8G")
	_        Handler = &RedisJWTHandler{}
)

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return h.client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid), "", h.rcExpiration).Err()
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	if err := h.SetRefreshToken(ctx, uid, ssid); err != nil {
		return err
	}
	return h.SetJWTToken(ctx, uid, ssid)
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
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

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		UserId:    uid,
		Ssid:      ssid,
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

func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)), // 7 天过期
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId    int64
	Ssid      string
	UserAgent string
}
