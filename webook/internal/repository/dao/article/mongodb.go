package article

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
	client  *mongo.Client
}

var (
	_                    ArticleDAO = (*MongoDBArticleDAO)(nil)
	ErrDocumentsNotFound            = mongo.ErrNoDocuments
)

func NewMongoDBArticleDAO(client *mongo.Client, node *snowflake.Node) ArticleDAO {
	mdb := client.Database("webook")
	return &MongoDBArticleDAO{
		node:    node,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		client:  client,
	}
}

func (m *MongoDBArticleDAO) ListPub(ctx context.Context, start time.Time, limit, offset int) ([]PublishedArticle, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	var articles []PublishedArticle
	filter := bson.A{
		bson.M{"update_time": bson.M{"$lt": start.UnixMilli()}}, // update_time < start.UnixMilli()
		bson.M{"status": domain.ArticleStatusPublished},
	}
	res, err := m.col.Find(ctx, bson.M{"$and": filter}, options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	err = res.All(ctx, &articles)
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, userId int64, limit, offset int) ([]Article, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	var articles []Article
	filter := bson.M{
		"author_id": userId,
	}
	sort := bson.M{"update_time": -1} // desc: -1; asc: 1
	res, err := m.col.Find(ctx, filter, options.Find().SetSort(sort).SetSkip(int64(offset)).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	err = res.All(ctx, &articles)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	var article Article
	filter := bson.M{
		"id": id,
	}
	res := m.col.FindOne(ctx, filter)
	if res.Err() != nil {
		return Article{}, res.Err()
	}
	err := res.Decode(&article)
	if err != nil {
		return Article{}, err
	}

	return article, nil
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	var article PublishedArticle
	filter := bson.M{
		"id": id,
	}
	res := m.col.FindOne(ctx, filter)
	if res.Err() != nil {
		return PublishedArticle{}, res.Err()
	}
	err := res.Decode(&article)
	if err != nil {
		return PublishedArticle{}, err
	}

	return article, nil
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	now := time.Now().UnixMilli()
	article.CreateTime = now
	article.UpdateTime = now
	article.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, &article)
	return article.Id, err
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, article Article) error {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	now := time.Now().UnixMilli()
	filter := bson.M{
		"id":        article.Id,
		"author_id": article.AuthorId,
	}
	set := bson.M{
		"$set": bson.M{
			"title":       article.Title,
			"content":     article.Content,
			"status":      article.Status,
			"update_time": now,
		},
	}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("ID或创作者错误")
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, article Article) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	id := article.Id
	sess, err := m.client.StartSession()
	if err != nil {
		return 0, err
	}
	defer func() {
		sess.EndSession(ctx)
	}()

	/** mongodb 事务方法
	 * 开启事务: sess.StartTransaction()
	 * 提交事务: sess.CommitTransaction()
	 * 回滚事务: sess.AbortTransaction()
	 */
	_, err = sess.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if id > 0 {
			err = m.UpdateById(ctx, article)
		} else {
			id, err = m.Insert(ctx, article)
		}
		if err != nil {
			return 0, err
		}
		article.Id = id
		now := time.Now().UnixMilli()
		article.UpdateTime = now
		// filter := bson.D{
		// 	bson.E{
		// 		Key:   "id",
		// 		Value: article.Id,
		// 	},
		// 	bson.E{
		// 		Key:   "author_id",
		// 		Value: article.AuthorId,
		// 	},
		// }
		filter := bson.M{
			"id":        id,
			"author_id": article.AuthorId,
		}
		set := bson.D{
			bson.E{
				Key:   "$set",
				Value: article,
			},
			bson.E{
				Key: "$setOnInsert",
				// Value: bson.D{
				// 	bson.E{
				// 		Key:   "create_time",
				// 		Value: now,
				// 	},
				// },
				Value: bson.M{"create_time": now},
			},
		}
		_, err = m.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
		return nil, err
	})
	return id, err
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, userId, id int64, status domain.ArticleStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// filter := bson.D{
	// 	bson.E{
	// 		Key:   "id",
	// 		Value: id,
	// 	},
	// 	bson.E{
	// 		Key:   "author_id",
	// 		Value: userId,
	// 	},
	// }
	filter := bson.M{
		"id":        id,
		"author_id": userId,
	}
	// sets := bson.D{
	// 	bson.E{
	// 		Key: "$set",
	// 		// Value: bson.D{
	// 		// 	bson.E{
	// 		// 		Key:   "status",
	// 		// 		Value: status.ToUint8(),
	// 		// 	},
	// 		// },
	// 		Value: bson.M{"status": status.ToUint8()},
	// 	},
	// }
	sets := bson.M{"$set": bson.M{"status": status.ToUint8()}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 0 {
		return errors.New("ID或创作者错误")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, sets)
	return err
}
