package web

import (
	"net/http"
	"strconv"
	"time"

	intrv1 "webook/api/proto/gen/intr/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/internal/domain"
	"webook/internal/errs"
	"webook/internal/service"
	"webook/internal/web/jwt"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"github.com/to404hanga/pkg404/ginx"
	"github.com/to404hanga/pkg404/logger"
	"golang.org/x/sync/errgroup"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc intrv1.InteractiveServiceClient
	reward  rewardv1.RewardServiceClient
	l       logger.Logger
	biz     string
}

func NewArticleHandler(l logger.Logger, svc service.ArticleService, intrSvc intrv1.InteractiveServiceClient, reward rewardv1.RewardServiceClient) *ArticleHandler {
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
	g.POST("/list", h.List)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true 就是点赞, false 就是不点赞
	pub.POST("/like", ginx.WrapBodyAndClaims(h.Like))
	pub.POST("/collect", ginx.WrapBodyAndClaims(h.Collect))
	pub.POST("/reward", ginx.WrapBodyAndClaims(h.Reward))
}

func (h *ArticleHandler) Reward(ctx *gin.Context, req ArticleRewardReq, uc jwt.UserClaims) (ginx.Result, error) {
	articleResp, err := h.svc.GetPubById(ctx, req.Id, uc.UserId)
	if err != nil {
		return ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	resp, err := h.reward.PreReward(ctx, &rewardv1.PreRewardRequest{
		Biz:       "article",
		BizId:     articleResp.Id,
		BizName:   articleResp.Title,
		TargetUid: articleResp.Author.Id,
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
func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleEditReq, uc jwt.UserClaims) (ginx.Result, error) {
	// 传入 ctx.Request.Context() 以确保 zipkin 正确读取
	id, err := h.svc.Save(ctx.Request.Context(), domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.UserId,
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

func (h *ArticleHandler) Publish(ctx *gin.Context, req PublishReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Publish(ctx.Request.Context(), domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.UserId,
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

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleWithdrawReq, uc jwt.UserClaims) (ginx.Result, error) {
	err := h.svc.Withdraw(ctx.Request.Context(), uc.UserId, req.Id)
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

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	// 我要不要检测一下？
	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx.Request.Context(), uc.UserId, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败", logger.Error(err), logger.Int("offset", page.Offset), logger.Int("limit", page.Limit), logger.Int64("UserId", uc.UserId))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:         src.Id,
				Title:      src.Title,
				Abstract:   src.Abstract(),
				AuthorId:   src.Author.Id,
				Status:     src.Status.ToUint8(),
				CreateTime: src.CreateTime.Format(time.DateTime),
				UpdateTime: src.UpdateTime.Format(time.DateTime),
			}
		}),
	})
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
	art, err := h.svc.GetById(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.ArticleInternalServerError,
		})
		h.l.Error("查询文章失败", logger.Int64("id", id), logger.Error(err))
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.Id != uc.UserId {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.ArticleInternalServerError,
		})
		h.l.Error("非法查询文章", logger.Int64("id", id), logger.Int64("UserId", uc.UserId))
		return
	}

	vo := ArticleVo{
		Id:         art.Id,
		Title:      art.Title,
		Content:    art.Content,
		AuthorId:   art.Author.Id,
		Status:     art.Status.ToUint8(),
		CreateTime: art.CreateTime.Format(time.DateTime),
		UpdateTime: art.UpdateTime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, ginx.Result{Data: vo})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
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

	var (
		eg   errgroup.Group
		art  domain.Article
		intr *intrv1.GetResponse
	)

	uc := ctx.MustGet("user").(jwt.UserClaims)
	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx.Request.Context(), id, uc.UserId)
		return er
	})
	eg.Go(func() error {
		var er error
		intr, er = h.intrSvc.Get(ctx.Request.Context(), &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   uc.UserId,
		})
		return er
	})

	// 等待结果
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: errs.ArticleInternalServerError,
		})
		h.l.Error("查询文章失败，系统错误", logger.Int64("aid", id), logger.Int64("UserId", uc.UserId), logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Data: ArticleVo{
			Id:         art.Id,
			Title:      art.Title,
			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			ReadCnt:    intr.GetIntr().GetReadCnt(),
			CollectCnt: intr.GetIntr().GetCollectCnt(),
			LikeCnt:    intr.GetIntr().GetLikeCnt(),
			Liked:      intr.GetIntr().GetLiked(),
			Collected:  intr.GetIntr().GetCollected(),
			Status:     art.Status.ToUint8(),
			CreateTime: art.CreateTime.Format(time.DateTime),
			UpdateTime: art.UpdateTime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req ArticleLikeReq, uc jwt.UserClaims) (ginx.Result, error) {
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

func (h *ArticleHandler) Collect(ctx *gin.Context, req ArticleCollectReq, uc jwt.UserClaims) (ginx.Result, error) {
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
