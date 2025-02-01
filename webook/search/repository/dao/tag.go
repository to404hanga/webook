package dao

import (
	"context"
	"encoding/json"

	"github.com/olivere/elastic/v7"
)

const TagIndexName = "tag_index"

type TagElasticSearchDAO struct {
	client *elastic.Client
}

var _ TagDAO = (*TagElasticSearchDAO)(nil)

func NewTagElasticSearchDAO(client *elastic.Client) TagDAO {
	return &TagElasticSearchDAO{client: client}
}

func (t *TagElasticSearchDAO) Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
	)
	resp, err := t.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var bt BizTags
		err := json.Unmarshal(hit.Source, &bt)
		if err != nil {
			return nil, err
		}
		res = append(res, bt.BizId)
	}
	return res, nil
}
