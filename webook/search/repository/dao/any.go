package dao

import (
	"context"

	"github.com/olivere/elastic/v7"
)

type AnyElasticSearchDAO struct {
	client *elastic.Client
}

var _ AnyDAO = (*AnyElasticSearchDAO)(nil)

func NewAnyElasticSearchDAO(client *elastic.Client) AnyDAO {
	return &AnyElasticSearchDAO{client: client}
}

func (a *AnyElasticSearchDAO) Input(ctx context.Context, index, docId, data string) error {
	_, err := a.client.Index().Index(index).Id(docId).BodyString(data).Do(ctx)
	return err
}
