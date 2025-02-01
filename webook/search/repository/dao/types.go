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
	Search(ctx context.Context, req SearchReq, keywords []string) ([]Article, error)
}

//go:generate mockgen -source=./types.go -destination=./mocks/tag.mock.go -package=daomocks TagDAO
type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}

//go:generate mockgen -source=./types.go -destination=./mocks/any.mock.go -package=daomocks AnyDAO
type AnyDAO interface {
	Input(ctx context.Context, index, docId, data string) error
	Delete(ctx context.Context, index, docId string) error
}

//go:generate mockgen -source=./types.go -destination=./mocks/like.mock.go -package=daomocks LikeDAO
type LikeDAO interface {
	Search(ctx context.Context, uid int64, biz string) ([]int64, error)
}

//go:generate mockgen -source=./types.go -destination=./mocks/collect.mock.go -package=daomocks CollectDAO
type CollectDAO interface {
	Search(ctx context.Context, uid int64, biz string) ([]int64, error)
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

type SearchReq struct {
	LikeIds    []int64
	TagIds     []int64
	CollectIds []int64
}

type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}
