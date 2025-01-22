package web

import (
	"net/http"
	"webook/payment/service/wechat"

	"github.com/gin-gonic/gin"
	"github.com/to404hanga/pkg404/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.Logger
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler, nativeSvc *wechat.NativePaymentService, l logger.Logger) *WechatHandler {
	return &WechatHandler{
		handler:   handler,
		nativeSvc: nativeSvc,
		l:         l,
	}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.Any("/pay/callback", h.HandleNative)
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) {
	// 用于接收解密后的数据
	transaction := new(payments.Transaction)
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		ctx.String(http.StatusBadRequest, "参数解析失败")
		h.l.Error("解析微信支付回调失败", logger.Error(err))
		return
	}
	err = h.nativeSvc.HandleCallBack(ctx, transaction)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统异常")
		h.l.Error("处理微信支付回调失败", logger.Error(err), logger.String("biz_trade_no", *transaction.OutTradeNo))
		return
	}
	ctx.String(http.StatusOK, "OK")
}
