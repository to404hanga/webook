package web

import (
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/errs"
	"webook/internal/service"
	myJwt "webook/internal/web/jwt"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/to404hanga/pkg404/ginx"
	"github.com/to404hanga/pkg404/logger"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	bizLogin             = "login"
)

// lsh040321@petalmail.com
// Aa#123456

type UserHandler struct {
	myJwt.Handler
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	logger         logger.Logger
	svc            service.UserService
	codeSvc        service.CodeService
}

var _ Handler = (*UserHandler)(nil)

func NewUserHandler(logger logger.Logger, svc service.UserService, codeSvc service.CodeService, handler myJwt.Handler) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		logger:         logger,
		svc:            svc,
		codeSvc:        codeSvc,
		Handler:        handler,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	users := server.Group("/users")
	{
		users.POST("/signup", ginx.WrapBody(h.SignUp))
		users.POST("/login", ginx.WrapBody(h.Login))
		users.POST("/logout", h.LogoutJWT)
		users.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))
		users.GET("/profile", ginx.WrapClaims(h.Profile))
		users.GET("/refresh_token", h.RefreshToken)

		users.POST("/login_sms/code/send", ginx.WrapBody(h.SendSMSLoginCode))
		users.POST("/login_sms", ginx.WrapBody(h.LoginSMS))
	}
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context, req SendSMSCodeReq) (ginx.Result, error) {
	if req.Phone == "" {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "请输入手机号码",
		}, nil
	}
	err := h.codeSvc.Send(ctx.Request.Context(), bizLogin, req.Phone)
	switch err {
	case nil:
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "发送成功",
		}, nil
	case service.ErrCodeSendTooMany:
		return ginx.Result{
			Code: errs.UserSmsSendTooMany,
			Msg:  "短信发送太频繁，请稍后再试",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	ok, err := h.codeSvc.Verify(ctx.Request.Context(), bizLogin, req.Phone, req.Code)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !ok {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "验证码错误，请重新输入",
		}, nil
	}
	user, err := h.svc.FindOrCreate(ctx.Request.Context(), req.Phone)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if err = h.SetLoginToken(ctx, user.Id); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "登陆成功",
	}, nil
}

func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	if isEmail, err := h.emailRexExp.MatchString(req.Email); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	} else if !isEmail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "非法邮箱格式",
		}, nil
	}

	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入的密码不一致",
		}, nil
	}

	if isPassword, err := h.passwordRexExp.MatchString(req.Password); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	} else if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码必须包含字母、数字、特殊字符，并且不少于八位",
		}, nil
	}

	err := h.svc.SignUp(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "注册成功",
		}, nil
	case service.ErrDuplicateUser:
		return ginx.Result{
			Code: errs.UserDuplicateEmail,
			Msg:  "该邮箱已被注册",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (h *UserHandler) Login(ctx *gin.Context, req LoginJWTReq) (ginx.Result, error) {
	user, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	switch err {
	case nil:
		if err = h.SetLoginToken(ctx, user.Id); err != nil {
			return ginx.Result{
				Code: errs.UserInternalServerError,
				Msg:  "系统错误",
			}, err
		}
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "登陆成功",
		}, nil
	case service.ErrInvalidUserOrPassword:
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "用户名或密码错误",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Msg:  "登出成功",
	})
}

func (h *UserHandler) Edit(ctx *gin.Context, req UserEditReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "非法的生日格式",
		}, err
	}
	if err = h.svc.EditNonSensitive(ctx.Request.Context(), domain.User{
		Id:       userClaims.UserId,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	}); err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "修改成功",
	}, nil
}

func (h *UserHandler) Profile(ctx *gin.Context, userClaims myJwt.UserClaims) (ginx.Result, error) {
	user, err := h.svc.FindById(ctx.Request.Context(), userClaims.UserId)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: User{
			Nickname: user.Nickname,
			Email:    user.Email,
			Birthday: user.Birthday.Format(time.DateOnly),
			AboutMe:  user.AboutMe,
		},
	}, nil
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	tokenStr := h.ExtractToken(ctx)
	var rc myJwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(t *jwt.Token) (interface{}, error) {
		return []byte(myJwt.RCJWTKey), nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err = h.CheckSession(ctx, rc.Ssid); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err = h.SetJWTToken(ctx, rc.Uid, rc.Ssid); err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
	})
}
