package dao

import (
	"context"
	"time"
	"webook/internal/repository/dao/article"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&article.Article{},
		&article.PublishedArticle{},
		&AsyncSms{},
		&Job{},
	)
}

func InitCollection(mdb *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	col := mdb.Collection("articles")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"author_id": 1},
		},
	})
	if err != nil {
		return err
	}
	liveCol := mdb.Collection("published_articles")
	_, err = liveCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.M{"id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"author_id": 1},
		},
	})
	return err
}
