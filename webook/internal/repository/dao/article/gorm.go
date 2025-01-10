package article

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormArticleDAO struct {
	db *gorm.DB
}

var _ ArticleDAO = (*GormArticleDAO)(nil)

func NewGormArticleDAO(db *gorm.DB) ArticleDAO {
	return &GormArticleDAO{
		db: db,
	}
}

func (dao *GormArticleDAO) ListPub(ctx context.Context, start time.Time, limit, offset int) ([]PublishedArticle, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	var articles []PublishedArticle
	err := dao.db.WithContext(ctx).Where("update_time < ? AND status = ?", start.UnixMilli(), domain.ArticleStatusPublished).Limit(limit).Offset(offset).Find(&articles).Error
	return articles, err
}

func (dao *GormArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var article PublishedArticle
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&article).Error
	return article, err
}

func (dao *GormArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var article Article
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&article).Error
	return article, err
}

func (dao *GormArticleDAO) GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]Article, error) {
	var articles []Article
	err := dao.db.WithContext(ctx).Where("author_id = ?", userId).Limit(limit).Offset(offset).Order("update_time DESC").Find(&articles).Error
	return articles, err
}

func (dao *GormArticleDAO) SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, userId).Updates(map[string]interface{}{
			"update_time": now,
			"status":      status,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("Id或创作者错误")
		}
		return tx.Model(&PublishedArticle{}).Where("id = ?", userId).Updates(map[string]interface{}{
			"update_time": now,
			"status":      status,
		}).Error
	})
}

func (dao *GormArticleDAO) Sync(ctx context.Context, article Article) (int64, error) {
	id := article.Id
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		txDao := NewGormArticleDAO(tx)
		if id > 0 {
			err = txDao.UpdateById(ctx, article)
		} else {
			id, err = txDao.Insert(ctx, article)
		}
		if err != nil {
			return err
		}
		article.Id = id
		now := time.Now().UnixMilli()
		pubArticle := PublishedArticle(article)
		pubArticle.CreateTime = now
		pubArticle.UpdateTime = now
		err = tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "id",
				},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":       pubArticle.Title,
				"content":     pubArticle.Content,
				"update_time": pubArticle.UpdateTime,
				"status":      pubArticle.Status,
			}),
		}).Create(&pubArticle).Error
		return err
	})
	return id, err
}

func (dao *GormArticleDAO) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.WithContext(ctx).Model(&article).Where("id = ? AND author_id = ?", article.Id, article.AuthorId).Updates(map[string]interface{}{
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

func (dao *GormArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.CreateTime = now
	article.UpdateTime = now
	err := dao.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}
