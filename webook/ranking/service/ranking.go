package service

import (
	"context"
	"math"
	"time"
	articlev1 "webook/api/proto/gen/article/v1"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/ranking/domain"
	"webook/ranking/repository"

	"github.com/to404hanga/pkg404/stl/queue"
	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -source=./ranking.go -package=svcmocks -destination=./mocks/ranking.mock.go RankingService
type RankingService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	intrSvc   intrv1.InteractiveServiceClient
	artSvc    articlev1.ArticleServiceClient
	repo      repository.RankingRepository
	batchSize int
	scoreFunc func(likeCnt int64, updateTime time.Time) float64
	n         int
}

func NewBatchRankingService(intrSvc intrv1.InteractiveServiceClient, artSvc articlev1.ArticleServiceClient, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, updateTime time.Time) float64 {
			duration := time.Since(updateTime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (s *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return s.repo.GetTopN(ctx)
}

func (s *BatchRankingService) TopN(ctx context.Context) error {
	articles, err := s.topN(ctx)
	if err != nil {
		return err
	}
	return s.repo.ReplaceTopN(ctx, articles)
}

func (s *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueueFunc(func(left, right Score) bool {
		return left.score < right.score
	})

	for {
		resp, err := s.artSvc.ListPub(ctx, &articlev1.ListPubRequest{
			StartTime: timestamppb.New(start),
			Offset:    int64(offset),
			Limit:     int64(s.batchSize),
		})
		if err != nil {
			return nil, err
		}
		articles := transform.SliceFromSlice[*articlev1.Article, domain.Article](resp.GetArticles(), func(idx int, a *articlev1.Article) domain.Article {
			return articleToDomain(a)
		})
		ids := make([]int64, 0, len(articles))
		for _, a := range articles {
			ids = append(ids, a.Id)
		}
		if len(articles) == 0 {
			break
		}
		intrResp, err := s.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article",
			Ids: ids,
		})
		if err != nil {
			return nil, err
		}
		intrMap := intrResp.GetIntrs()
		for _, a := range articles {
			intr := intrMap[a.Id]
			score := s.scoreFunc(intr.LikeCnt, a.UpdateTime)
			ele := Score{
				score: score,
				art:   a,
			}
			topN.Push(ele)
			if topN.Len() > s.n {
				min := topN.Pop()
				if min.score < score {
					topN.Push(ele)
				} else {
					topN.Push(min)
				}
			}
		}
		// 优化: 本批次最后一条的时间在七天前，直接返回
		if len(articles) < s.batchSize || articles[len(articles)-1].UpdateTime.Before(ddl) {
			break
		}
		offset += len(articles)
	}

	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele := topN.Pop()
		res[i] = ele.art
	}
	return res, nil
}

func articleToDomain(article *articlev1.Article) domain.Article {
	domainArticle := domain.Article{}
	if article != nil {
		domainArticle.Id = article.GetId()
		domainArticle.Title = article.GetTitle()
		domainArticle.Status = domain.ArticleStatus(article.Status)
		domainArticle.Content = article.Content
		domainArticle.Author = domain.Author{
			Id:   article.GetAuthor().GetId(),
			Name: article.GetAuthor().GetName(),
		}
		domainArticle.CreateTime = article.CreateTime.AsTime()
		domainArticle.UpdateTime = article.UpdateTime.AsTime()
	}
	return domainArticle
}
