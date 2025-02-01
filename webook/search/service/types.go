package service

import (
	"context"
	"webook/search/domain"
)

//go:generate mockgen -source=./types.go -destination=./mocks/sync.mock.go -package=svcmocks SyncService
type SyncService interface {
	InputArticle(ctx context.Context, article domain.Article) error
	InputUser(ctx context.Context, user domain.User) error
	InputAny(ctx context.Context, indexName, docId, data string) error
	Delete(ctx context.Context, indexName, docId string) error
}

//go:generate mockgen -source=./types.go -destination=./mocks/search.mock.go -package=svcmocks SearchService
type SearchService interface {
	Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error)
}
