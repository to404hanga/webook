package dao

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/olivere/elastic/v7"
	"github.com/to404hanga/pkg404/stl/transform"
)

const ArticleIndexName = "article_index"

const (
	ArticleStatusPublished = 2
)

type ArticleElasticSearchDAO struct {
	client *elastic.Client
}

var _ ArticleDAO = (*ArticleElasticSearchDAO)(nil)

func NewArticleElasticSearchDAO(client *elastic.Client) ArticleDAO {
	return &ArticleElasticSearchDAO{client: client}
}

func (a *ArticleElasticSearchDAO) InputArticle(ctx context.Context, article Article) error {
	_, err := a.client.Index().Index(ArticleIndexName).BodyJson(article).Do(ctx)
	return err
}

func (a *ArticleElasticSearchDAO) Search(ctx context.Context, req SearchReq, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	status := elastic.NewTermQuery("status", ArticleStatusPublished)

	title := elastic.NewMatchQuery("title", queryString)
	content := elastic.NewMatchQuery("content", queryString)

	tag := elastic.NewTermsQuery("id", transform.SliceFromSlice[int64, any](req.TagIds, func(i int64) any {
		return i
	})...).Boost(2)
	collect := elastic.NewTermsQuery("id", transform.SliceFromSlice[int64, any](req.CollectIds, func(i int64) any {
		return i
	})...).Boost(4)
	like := elastic.NewTermsQuery("id", transform.SliceFromSlice[int64, any](req.LikeIds, func(i int64) any {
		return i
	})...).Boost(2)

	or := elastic.NewBoolQuery().Should(title, content, tag, collect, like)
	query := elastic.NewBoolQuery().Must(status, or)
	sort := elastic.NewFieldSort("id").Desc()
	scoreSort := elastic.NewFieldSort("_score").Desc()
	resp, err := a.client.Search(ArticleIndexName).SortBy(scoreSort, sort).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var article Article
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			return nil, err
		}
		res = append(res, article)
	}
	return res, nil
}
