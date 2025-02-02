package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ArticleGormAuthorDAO struct {
	db *gorm.DB
}

var _ ArticleAuthorDAO = (*ArticleGormAuthorDAO)(nil)

func NewArticleGormAuthorDAO(db *gorm.DB) ArticleAuthorDAO {
	return &ArticleGormAuthorDAO{db: db}
}

func (a *ArticleGormAuthorDAO) Create(ctx context.Context, article Article) (int64, error) {
	err := a.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (a *ArticleGormAuthorDAO) Update(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&article).Where("id = ? AND author_id = ?", article.Id, article.AuthorId).Updates(map[string]any{
		"title":       article.Title,
		"content":     article.Content,
		"status":      article.Status,
		"update_time": now,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("Id或创作者错误")
	}
	return nil
}
