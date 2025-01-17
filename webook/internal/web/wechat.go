package web

import (
	"fmt"
	"net/http"
	"webook/internal/errs"
	"webook/internal/service"
	"webook/internal/service/oauth2/wechat"
	myJwt "webook/internal/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/to404hanga/pkg404/ginx"
)

type OAuth2WechatHandler struct {
	myJwt.Handler
	svc             wechat.Service
	userSvc         service.UserService
	key             []byte
	stateCookieName string
}

var _ Handler = (*OAuth2WechatHandler)(nil)

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, handler myJwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("EZUAnsruwZew6sVuEXUhRjr7p9INoqnw2EkrFr47oH2Q9D99dESoa3LTVklrKP22"),
		stateCookieName: "jwt-state",
		Handler:         handler,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	wechat := server.Group("/oauth2/wechat")
	{
		wechat.GET("/authurl", h.Auth2URL)
		wechat.Any("/callback", h.Callback)
	}
}

func (h *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	state := uuid.New()
	value, err := h.svc.Auth2URL(ctx.Request.Context(), state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "构造跳转URL失败",
			Code: errs.WechatInternalServerError,
		})
		return
	}
	err = h.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.WechatInternalServerError,
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: value,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "非法请求",
			Code: errs.WechatInvalidInput,
		})
		return
	}
	code := ctx.Query("code")
	wechatInfo, err := h.svc.VerifyCode(ctx.Request.Context(), code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "授权失败",
			Code: errs.WechatAuthorizeFailed,
		})
		return
	}
	user, err := h.userSvc.FindOrCreateByWechat(ctx.Request.Context(), wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.WechatInternalServerError,
		})
		return
	}

	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	ctx.JSON(http.StatusOK, ginx.Result{
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
