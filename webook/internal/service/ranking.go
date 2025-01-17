package service

import (
	"context"
	"math"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/internal/domain"
	"webook/internal/repository"

	"github.com/ecodeclub/ekit/queue"
)

//go:generate mockgen -source=./ranking.go -package=svcmocks -destination=./mocks/ranking.mock.go RankingService
type RankingService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	intrSvc   intrv1.InteractiveServiceClient
	artSvc    ArticleService
	repo      repository.RankingRepository
	batchSize int
	scoreFunc func(likeCnt int64, updateTime time.Time) float64
	n         int
}

func NewBatchRankingService(intrSvc intrv1.InteractiveServiceClient, artSvc ArticleService, repo repository.RankingRepository) RankingService {
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
	panic(articles)
}

func (s *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue(s.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score < dst.score {
			return -1
		}
		return 0
	})

	for {
		articles, err := s.artSvc.ListPub(ctx, start, s.batchSize, offset)
		if err != nil {
			return nil, err
		}
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
			err = topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				min, _ := topN.Dequeue()
				if min.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(min)
				}
			}
		}
		offset += len(articles)
		// 优化: 本批次最后一条的时间在七天前，直接返回
		if len(articles) < s.batchSize || articles[len(articles)-1].UpdateTime.Before(ddl) {
			break
		}
	}

	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
