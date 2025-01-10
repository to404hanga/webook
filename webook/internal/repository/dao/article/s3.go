package article

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"time"
	"webook/internal/domain"

	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type S3ArticleDAO struct {
	GormArticleDAO
	oss *s3.S3
}

var _ ArticleDAO = (*S3ArticleDAO)(nil)

type PublishedArticleV2 struct {
	Id         int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title      string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	AuthorId   int64  `gorm:"index" bson:"author_id,omitempty"`
	Status     uint8  `bson:"status,omitempty"`
	CreateTime int64  `bson:"create_time,omitempty"`
	UpdateTime int64  `bson:"update_time,omitempty"`
}

func NewS3ArticleDAO(db *gorm.DB, oss *s3.S3) ArticleDAO {
	return &S3ArticleDAO{
		GormArticleDAO: GormArticleDAO{
			db: db,
		},
		oss: oss,
	}
}

func (dao *S3ArticleDAO) SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, userId).Updates(map[string]interface{}{
			"update_time": now,
			"status":      status,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID或创作者错误")
		}
		return tx.Model(&PublishedArticle{}).Where("id = ?", id).Updates(map[string]interface{}{
			"update_time": now,
			"status":      status,
		}).Error
	})
	if err != nil {
		return err
	}
	if status == domain.ArticleStatusPrivate {
		bucket := "webook-1314583317"
		key := strconv.FormatInt(id, 10)
		_, err = dao.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
	}
	return err
}

func (dao *S3ArticleDAO) Sync(ctx context.Context, article Article) (int64, error) {
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
		pubArticle := PublishedArticleV2{
			Id:         id,
			Title:      article.Title,
			AuthorId:   article.AuthorId,
			Status:     article.Status,
			CreateTime: now,
			UpdateTime: now,
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{
					Name: "id",
				},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":       pubArticle.Title,
				"status":      pubArticle.Status,
				"update_time": now,
			}),
		}).Create(&pubArticle).Error
	})
	if err != nil {
		return 0, err
	}
	bucket := "webook-1314583317"
	key := strconv.FormatInt(id, 10)
	contentType := "text/plain;charset=utf-8"
	_, err = dao.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        bytes.NewReader([]byte(article.Content)),
		ContentType: &contentType,
	})
	return id, err
}
