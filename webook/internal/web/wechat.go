package web

import (
	"fmt"
	"net/http"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
)

type OAuth2WechatHandler struct {
	jwtHandler
	svc             wechat.Service
	userSvc         service.UserService
	key             []byte
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("EZUAnsruwZew6sVuEXUhRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP22"),
		stateCookieName: "jwt-state",
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	{
		g.GET("/authurl", h.Auth2URL)
		g.Any("/callback", h.Callback)
	}
}

func (h *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	state := uuid.New()
	value, err := h.svc.Auth2URL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 500,
		})
		return
	}
	err = h.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: http.StatusInternalServerError,
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: value,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: http.StatusBadRequest,
		})
		return
	}
	code := ctx.Query("code")
	wechatInfo, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "授权失败",
			Code: http.StatusUnauthorized,
		})
		return
	}
	user, err := h.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: http.StatusInternalServerError,
		})
		return
	}
	h.setJWTToken(ctx, user.Id)
	ctx.JSON(http.StatusOK, Result{
		Code: http.StatusOK,
		Msg:  "OK",
	})
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	cookie, err := ctx.Cookie(h.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得 cookie: %v", err)
	}
	var stateClaims StateClaims
	_, err = jwt.ParseWithClaims(cookie, &stateClaims, func(t *jwt.Token) (interface{}, error) {
		return h.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %v", err)
	}
	if state != stateClaims.State {
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.key)
	if err != nil {
		return err
	}
	ctx.SetCookie(h.stateCookieName, tokenStr, 600, "/oauth2/wechat/callback", "", false, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
