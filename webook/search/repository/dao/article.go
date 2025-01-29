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

func (a *ArticleElasticSearchDAO) Search(ctx context.Context, articleIds []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	status := elastic.NewTermQuery("status", ArticleStatusPublished)

	title := elastic.NewMatchQuery("title", queryString)
	content := elastic.NewMatchQuery("content", queryString)
	tag := elastic.NewTermQuery("id", transform.SliceFromSlice[int64, any](articleIds, func(i int64) any {
		return i
	})).Boost(2)

	or := elastic.NewBoolQuery().Should(title, content, tag)
	query := elastic.NewBoolQuery().Must(status, or)
	resp, err := a.client.Search(ArticleIndexName).Query(query).Do(ctx)
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
