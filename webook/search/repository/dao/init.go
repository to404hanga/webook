package dao

import (
	"context"
	_ "embed"
	"time"

	"github.com/olivere/elastic/v7"
	"golang.org/x/sync/errgroup"
)

var (
	//go:embed json/user_index.json
	userIndex string
	//go:embed json/article_index.json
	articleIndex string
	//go:embed json/tag_index.json
	tagIndex string
)

func InitES(client *elastic.Client) error {
	const timeout = time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var eg errgroup.Group
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, UserIndexName, userIndex)
	})
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, ArticleIndexName, articleIndex)
	})
	eg.Go(func() error {
		return tryCreateIndex(ctx, client, TagIndexName, tagIndex)
	})
	return eg.Wait()
}

func tryCreateIndex(ctx context.Context, client *elastic.Client, indexName, indexConfig string) error {
	ok, err := client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = client.CreateIndex(indexName).Body(indexConfig).Do(ctx)
	return err
}
