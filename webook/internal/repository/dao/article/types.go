package article

import (
	"context"
	"time"
	"webook/internal/domain"
)

type ArticleDAO interface {
	UpdateById(ctx context.Context, article Article) error
	Insert(ctx context.Context, article Article) (int64, error)
	ListPub(ctx context.Context, start time.Time, limit, offset int) ([]PublishedArticle, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]Article, error)
	SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error
	Sync(ctx context.Context, article Article) (int64, error)
}

type Article struct {
	Id         int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title      string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content    string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorId   int64  `gorm:"index" bson:"author_id,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
	CreateTime int64  `bson:"create_time,omitempty"`
	UpdateTime int64  `bson:"update_time,omitempty"`
}

type PublishedArticle Article
