package dao

import "context"

//go:generate mockgen -source=./types.go -destination=./mocks/user.mock.go -package=daomocks UserDAO
type UserDAO interface {
	InputUser(ctx context.Context, user User) error
	Search(ctx context.Context, keywords []string) ([]User, error)
}

//go:generate mockgen -source=./types.go -destination=./mocks/article.mock.go -package=daomocks ArticleDAO
type ArticleDAO interface {
	InputArticle(ctx context.Context, article Article) error
	Search(ctx context.Context, articleIds []int64, keywords []string) ([]Article, error)
}

//go:generate mockgen -source=./types.go -destination=./mocks/tag.mock.go -package=daomocks TagDAO
type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}

//go:generate mockgen -source=./types.go -destination=./mocks/any.mock.go -package=daomocks AnyDAO
type AnyDAO interface {
	Input(ctx context.Context, index, docId, data string) error
}

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

type Article struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Status  int32    `json:"status"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}
