package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/internal/domain"
	"webook/internal/service"
	myJwt "webook/internal/web/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc intrv1.InteractiveServiceClient
	reward  rewardv1.RewardServiceClient
	l       logger.Logger
	biz     string
}

func NewArticleHandler(svc service.ArticleService, intrSvc intrv1.InteractiveServiceClient, reward rewardv1.RewardServiceClient, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		reward:  reward,
		l:       l,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	articles := server.Group("/articles")
	{
		articles.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))
		articles.POST("/publish", ginx.WrapBodyAndClaims(h.Publish))
		articles.POST("/withdraw", ginx.WrapBodyAndClaims(h.Withdraw))
		articles.GET("/detail/:id", h.Detail)
		articles.POST("/list", h.List)
		publishArticles := articles.Group("/pub")
		{
			publishArticles.GET("/:id", h.PubDetail)
			publishArticles.POST("/like", ginx.WrapBodyAndClaims(h.Like))
			publishArticles.POST("/collect", ginx.WrapBodyAndClaims(h.Collect))
			publishArticles.POST("/reward", ginx.WrapBodyAndClaims(h.Reward))
		}
	}
}

func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleEditReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: userClaims.UserId,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req PublishReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: userClaims.UserId,
		},
	})
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: http.StatusInternalServerError,
		}, fmt.Errorf("发表文章失败 aid %d, uid %d %w", userClaims.UserId, req.Id, err)
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: id,
	}, nil
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleWithdrawReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	if err := h.svc.Withdraw(ctx, userClaims.UserId, req.Id); err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: http.StatusInternalServerError,
		}, err
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
	userClaims := ctx.MustGet("user").(myJwt.UserClaims)
	articles, err := h.svc.GetByAuthor(ctx, userClaims.UserId, page.Limit, page.Offset)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败", logger.Error(err), logger.Int("limit", page.Limit), logger.Int("offset", page.Offset), logger.Int64("uid", userClaims.UserId))
		return
	}
	res := make([]ArticleVo, 0, len(articles))
	for _, article := range articles {
		res = append(res, ArticleVo{
			Id:         article.Id,
			Title:      article.Title,
			Abstract:   article.Abstract(),
			AuthorId:   article.Author.Id,
			Status:     article.Status.ToUint8(),
			CreateTime: article.CreateTime.Format(time.DateTime),
			UpdateTime: article.UpdateTime.Format(time.DateTime),
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Data: res,
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
		})
		h.l.Warn("查询文章失败, id 格式错误", logger.String("id", idStr), logger.Error(err))
		return
	}
	articles, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败", logger.Int64("id", id), logger.Error(err))
		return
	}
	userClaims := ctx.MustGet("user").(myJwt.UserClaims)
	if articles.Author.Id != userClaims.UserId {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusForbidden,
			Msg:  "无权限",
		})
		h.l.Error("非法查询文章", logger.Int64("id", id), logger.Int64("uid", userClaims.UserId))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Data: ArticleVo{
			Id:         articles.Id,
			Title:      articles.Title,
			Content:    articles.Content,
			AuthorId:   articles.Author.Id,
			Status:     articles.Status.ToUint8(),
			CreateTime: articles.CreateTime.Format(time.DateTime),
			UpdateTime: articles.UpdateTime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
		})
		h.l.Warn("查询文章失败, id 格式错误", logger.String("id", idStr), logger.Error(err))
		return
	}

	var (
		eg      errgroup.Group
		article domain.Article
		intr    *intrv1.GetResponse
	)

	userClaims := ctx.MustGet("user").(myJwt.UserClaims)
	eg.Go(func() error {
		var er error
		article, er = h.svc.GetPubById(ctx, id, userClaims.UserId)
		return er
	})
	eg.Go(func() error {
		var er error
		intr, er = h.intrSvc.Get(ctx, &intrv1.GetRequest{
			Biz:   h.biz,
			BizId: id,
			Uid:   userClaims.UserId,
		})
		return er
	})
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: http.StatusInternalServerError,
		})
		h.l.Error("查询文章失败，系统错误", logger.Int64("aid", id), logger.Int64("uid", userClaims.UserId), logger.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: ArticleVo{
			Id:         article.Id,
			Title:      article.Title,
			Content:    article.Content,
			AuthorId:   article.Author.Id,
			AuthorName: article.Author.Name,
			ReadCnt:    intr.Intr.ReadCnt,
			CollectCnt: intr.Intr.CollectCnt,
			LikeCnt:    intr.Intr.LikeCnt,
			Collected:  intr.Intr.Collected,
			Status:     article.Status.ToUint8(),
			CreateTime: article.CreateTime.Format(time.DateTime),
			UpdateTime: article.UpdateTime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req ArticleLikeReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		_, err = h.intrSvc.Like(ctx, &intrv1.LikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   userClaims.UserId,
		})
	} else {
		_, err = h.intrSvc.CancelLike(ctx, &intrv1.CancelLikeRequest{
			Biz:   h.biz,
			BizId: req.Id,
			Uid:   userClaims.UserId,
		})
	}
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
	}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req ArticleCollectReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	_, err := h.intrSvc.Collect(ctx, &intrv1.CollectRequest{
		Biz:   h.biz,
		BizId: req.Id,
		Uid:   userClaims.UserId,
		Cid:   req.Cid,
	})
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
	}, nil
}

func (h *ArticleHandler) Reward(ctx *gin.Context, req ArticleRewardReq, userClaims myJwt.UserClaims) (ginx.Result, error) {
	resp, err := h.svc.GetPubById(ctx, req.Id, userClaims.UserId)
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	res, err := h.reward.PreReward(ctx, &rewardv1.PreRewardRequest{
		Biz:       h.biz,
		BizId:     resp.Id,
		BizName:   resp.Title,
		TargetUid: resp.Author.Id,
		Uid:       userClaims.UserId,
		Amt:       req.Amt,
	})
	if err != nil {
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "OK",
		Data: map[string]interface{}{
			"codeURL": res.CodeUrl,
			"rid":     res.Rid,
		},
	}, nil
}
