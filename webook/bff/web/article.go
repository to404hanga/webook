package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	articlev1 "webook/api/proto/gen/article/v1"
	intrv1 "webook/api/proto/gen/intr/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/bff/web/jwt"
	"webook/errs"

	"github.com/gin-gonic/gin"
	"github.com/to404hanga/pkg404/ginx"
	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/stl/transform"
	"golang.org/x/sync/errgroup"
)

type ArticleHandler struct {
	svc     articlev1.ArticleServiceClient
	intrSvc intrv1.InteractiveServiceClient
	reward  rewardv1.RewardServiceClient
	l       logger.Logger
	biz     string
}

func NewArticleHandler(l logger.Logger, svc articlev1.ArticleServiceClient, intrSvc intrv1.InteractiveServiceClient, reward rewardv1.RewardServiceClient) *ArticleHandler {
	return &ArticleHandler{
		l:       l,
		svc:     svc,
		intrSvc: intrSvc,
		reward:  reward,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")

	g.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))
	g.POST("/publish", ginx.WrapBodyAndClaims(h.Publish))
	g.POST("/withdraw", ginx.WrapBodyAndClaims(h.Withdraw))

	// 创作者接口
	g.GET("/detail/:id", h.Detail)
	g.POST("/list", ginx.WrapBodyAndClaims(h.List))

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapClaims(h.PubDetail))
	// 传入一个参数，true 就是点赞, false 就是不点赞
	pub.POST("/like", ginx.WrapBodyAndClaims(h.Like))
	pub.POST("/collect", ginx.WrapBodyAndClaims(h.Collect))
	pub.POST("/reward", ginx.WrapBodyAndClaims(h.Reward))
}

func (h *ArticleHandler) Reward(ctx *gin.Context, req RewardReq, uc jwt.UserClaims) (ginx.Result, error) {
	articleResp, err := h.svc.GetPubById(ctx, &articlev1.GetPubByIdRequest{
		Id:  req.Id,
		Uid: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	resp, err := h.reward.PreReward(ctx, &rewardv1.PreRewardRequest{
		Biz:       "article",
		BizId:     articleResp.GetArticle().GetId(),
		BizName:   articleResp.GetArticle().GetTitle(),
		TargetUid: articleResp.GetArticle().GetAuthor().GetId(),
		Uid:       uc.UserId,
		Amt:       req.Amt,
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Data: map[string]any{
			"codeURL": resp.CodeUrl,
			"rid":     resp.Rid,
		},
	}, nil
}

// Edit 接收 Article 输入，返回一个 ID，文章的 ID
func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc jwt.UserClaims) (ginx.Result, error) {
	// 传入 ctx.Request.Context() 以确保 zipkin 正确读取
	id, err := h.svc.Save(ctx.Request.Context(), &articlev1.SaveRequest{
		Article: &articlev1.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: &articlev1.Author{
				Id: uc.UserId,
			},
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Data: id,
	}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Publish(ctx.Request.Context(), &articlev1.PublishRequest{
		Article: &articlev1.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: &articlev1.Author{
				Id: uc.UserId,
			},
		},
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Data: id,
	}, nil
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleReq, uc jwt.UserClaims) (ginx.Result, error) {
	_, err := h.svc.Withdraw(ctx.Request.Context(), &articlev1.WithdrawRequest{
		Id:  req.Id,
		Uid: uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: errs.ArticleInternalServerError,
		}, nil
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
	}, nil
}

func (a *ArticleHandler) List(ctx *gin.Context, req ListReq, usr jwt.UserClaims) (ginx.Result, error) {
	if req.Limit > 100 {
		a.l.Error("获得用户会话信息失败，LIMIT过大")
		return ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "请求有误",
		}, nil
	}

	arts, err := a.svc.List(ctx, &articlev1.ListRequest{
		Author: usr.UserId,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		a.l.Error("获得用户会话信息失败")
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Data: transform.SliceFromSlice[*articlev1.Article, ArticleVo](arts.Articles, func(idx int, src *articlev1.Article) ArticleVo {
			return ArticleVo{
				Id:         src.Id,
				Title:      src.Title,
				Abstract:   src.Abstract,
				Status:     src.Status,
				CreateTime: src.CreateTime.AsTime().Format(time.DateTime),
				UpdateTime: src.UpdateTime.AsTime().Format(time.DateTime),
			}
		}),
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "id 参数错误",
			Code: errs.ArticleInvalidInput,
		})
		h.l.Warn("查询文章失败，id 格式不对", logger.String("id", idStr), logger.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx.Request.Context(), &articlev1.GetByIdRequest{
		Id: id,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.ArticleInternalServerError,
		})
		h.l.Error("查询文章失败", logger.Int64("id", id), logger.Error(err))
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	article := art.GetArticle()
	if article.GetAuthor().GetId() != uc.UserId {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.ArticleInternalServerError,
		})
		h.l.Error("非法查询文章", logger.Int64("id", id), logger.Int64("UserId", uc.UserId))
		return
	}

	vo := ArticleVo{
		Id:         article.GetId(),
		Title:      article.GetTitle(),
		Content:    article.GetContent(),
		Author:     article.GetAuthor().GetName(),
		Status:     article.GetStatus(),
		CreateTime: article.GetCreateTime().String(),
		UpdateTime: article.GetUpdateTime().String(),
	}
	ctx.JSON(http.StatusOK, ginx.Result{Data: vo})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context, uc jwt.UserClaims) (ginx.Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.l.Error("前端输入的 ID 不对", logger.Error(err))
		return ginx.Result{
			Code: errs.ArticleInvalidInput,
			Msg:  "参数错误",
		}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idStr, err)
	}

	// 使用 error group 来同时查询数据
	var (
		eg       errgroup.Group
		artResp  *articlev1.GetPubByIdResponse
		intrResp *intrv1.GetResponse
	)
	eg.Go(func() error {
		var er error
		artResp, er = h.svc.GetPubById(ctx, &articlev1.GetPubByIdRequest{
			Id: id, Uid: uc.UserId,
		})
		return er
	})

	eg.Go(func() error {
		var er error
		intrResp, er = h.intrSvc.Get(ctx, &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   uc.UserId,
		})
		return er
	})

	err = eg.Wait()

	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("获取文章信息失败 %w", err)
	}

	art := artResp.GetArticle()
	intr := intrResp.Intr
	return ginx.Result{
		Data: ArticleVo{
			Id:         art.Id,
			Title:      art.Title,
			Status:     art.Status,
			Content:    art.Content,
			Author:     art.Author.Name,
			CreateTime: art.CreateTime.AsTime().Format(time.DateTime),
			UpdateTime: art.UpdateTime.AsTime().Format(time.DateTime),
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
			LikeCnt:    intr.LikeCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,
		},
	}, nil
}

func (h *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		// 点赞
		_, err = h.intrSvc.Like(ctx.Request.Context(), &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	} else {
		// 取消点赞
		_, err = h.intrSvc.CancelLike(ctx.Request.Context(), &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   uc.UserId,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
	}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc jwt.UserClaims) (ginx.Result, error) {
	_, err := h.intrSvc.Collect(ctx.Request.Context(), &intrv1.CollectRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Cid:   req.Cid,
		Uid:   uc.UserId,
	})
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
	}, nil
}
