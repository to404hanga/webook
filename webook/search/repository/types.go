package repository

import (
	"context"
	"webook/search/domain"
)

//go:generate mockgen -source=./types.go -package=repomocks -destination=./mocks/user.mock.go UserRepository
type UserRepository interface {
	InputUser(ctx context.Context, msg domain.User) error
	SearchUser(ctx context.Context, keywords []string) ([]domain.User, error)
}

//go:generate mockgen -source=./types.go -package=repomocks -destination=./mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	InputArticle(ctx context.Context, msg domain.Article) error
	SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error)
}

//go:generate mockgen -source=./types.go -package=repomocks -destination=./mocks/any.mock.go AnyRepository
type AnyRepository interface {
	Input(ctx context.Context, index, docId, data string) error
	Delete(ctx context.Context, index, docId string) error
}
