package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleGormReaderDAO struct {
	db *gorm.DB
}

var _ ArticleReaderDAO = (*ArticleGormReaderDAO)(nil)

func NewArticleGormReaderDAO(db *gorm.DB) ArticleReaderDAO {
	return &ArticleGormReaderDAO{db: db}
}

func (a *ArticleGormReaderDAO) Upsert(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.UpdateTime = now
	article.CreateTime = now
	return a.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"title":       article.Title,
			"content":     article.Content,
			"status":      article.Status,
			"update_time": article.UpdateTime,
		}),
	}).Create(&article).Error
}
