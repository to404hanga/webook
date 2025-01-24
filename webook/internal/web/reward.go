package web

import (
	"net/http"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/internal/errs"
	"webook/internal/web/jwt"

	"github.com/gin-gonic/gin"
	"github.com/to404hanga/pkg404/ginx"
)

type RewardHandler struct {
	client rewardv1.RewardServiceClient
}

func NewRewardHandler(client rewardv1.RewardServiceClient) *RewardHandler {
	return &RewardHandler{client: client}
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/reward")
	{
		rg.POST("/detail", ginx.WrapBodyAndClaims(h.GetReward))
	}
}

type GetRewardReq struct {
	Rid int64
}

func (h *RewardHandler) GetReward(ctx *gin.Context, req GetRewardReq, uc jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetReward(ctx, &rewardv1.GetRewardRequest{
		Rid: req.Rid,
		Uid: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: errs.RewardInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Data: resp.GetStatus().String(),
	}, nil
}
