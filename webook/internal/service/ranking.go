package service

import (
	"context"
	"math"
	"time"
	"webook/internal/domain"

	"github.com/ecodeclub/ekit/queue"
)

//go:generate mockgen -source=./ranking.go -package=svcmocks -destination=./mocks/ranking.mock.go RankingService
type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	intrSvc   InteractiveService
	artSvc    ArticleService
	batchSize int
	scoreFunc func(likeCnt int64, updateTime time.Time) float64
	n         int
}

func NewBatchRankingService(intrSvc InteractiveService, artSvc ArticleService) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, updateTime time.Time) float64 {
			duration := time.Since(updateTime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
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
		intrMap, err := s.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
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
