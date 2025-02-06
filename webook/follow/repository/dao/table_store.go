package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"gorm.io/gorm"
)

const FollowRelationTableName = "follow_relations"

var ErrFollowNotFound = gorm.ErrRecordNotFound

type TableStoreFollowRelationDAO struct {
	client *tablestore.TableStoreClient
}

var _ FollowRelationDAO = (*TableStoreFollowRelationDAO)(nil)

func NewTableStoreFollowRelationDAO(client *tablestore.TableStoreClient) FollowRelationDAO {
	return &TableStoreFollowRelationDAO{client: client}
}

func (t *TableStoreFollowRelationDAO) FollowRelationList(ctx context.Context, follower int64, limit, offset int) ([]FollowRelation, error) {
	req := &tablestore.SQLQueryRequest{Query: fmt.Sprintf("SELECT id,follower,followee FROM %s WHERE follower=%d AND status=%d OFFSET %d LIMIT %d", FollowRelationTableName, follower, FollowRelationStatusActive, offset, limit)}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return nil, err
	}
	resultSet := resp.ResultSet
	followRelations := make([]FollowRelation, 0, limit)
	for resultSet.HasNext() {
		row := resultSet.Next()
		followRelation := FollowRelation{}
		followRelation.Follower, _ = row.GetInt64ByName("follower")
		followRelation.Followee, _ = row.GetInt64ByName("followee")
		followRelations = append(followRelations, followRelation)
	}
	return followRelations, nil
}

func (t *TableStoreFollowRelationDAO) FansList(ctx context.Context, followee int64, limit, offset int) ([]FollowRelation, error) {
	req := &tablestore.SQLQueryRequest{Query: fmt.Sprintf("SELECT id,follower,followee FROM %s WHERE followee=%d AND status=%d OFFSET %d LIMIT %d", FollowRelationTableName, followee, FollowRelationStatusActive, offset, limit)}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return nil, err
	}
	resultSet := resp.ResultSet
	followRelations := make([]FollowRelation, 0, limit)
	for resultSet.HasNext() {
		row := resultSet.Next()
		followRelation := FollowRelation{}
		followRelation.Follower, _ = row.GetInt64ByName("follower")
		followRelation.Followee, _ = row.GetInt64ByName("followee")
		followRelations = append(followRelations, followRelation)
	}
	return followRelations, nil
}

func (t *TableStoreFollowRelationDAO) FollowRelationDetail(ctx context.Context, follower, followee int64) (FollowRelation, error) {
	req := &tablestore.SQLQueryRequest{Query: fmt.Sprintf("SELECT id,follower,followee FROM %s WHERE follower=%d AND followee=%d AND status=%d", FollowRelationTableName, follower, followee, FollowRelationStatusActive)}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return FollowRelation{}, err
	}
	resultSet := resp.ResultSet
	if resultSet.HasNext() {
		row := resultSet.Next()
		return t.rowToEntity(row), nil
	}
	return FollowRelation{}, ErrFollowNotFound
}

func (t *TableStoreFollowRelationDAO) CreateFollowRelation(ctx context.Context, c FollowRelation) error {
	req := new(tablestore.UpdateRowRequest)
	change := new(tablestore.UpdateRowChange)
	change.TableName = FollowRelationTableName
	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("follower", c.Follower)
	putPk.AddPrimaryKeyColumn("followee", c.Followee)
	change.PrimaryKey = putPk
	now := time.Now().Unix()
	change.PutColumn("status", int64(FollowRelationStatusActive))
	change.PutColumn("create_time", now)
	change.PutColumn("update_time", now)
	change.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	req.UpdateRowChange = change
	_, err := t.client.UpdateRow(req)
	return err
}

func (t *TableStoreFollowRelationDAO) UpdateStatus(ctx context.Context, followee, follower int64, status uint8) error {
	condition := tablestore.NewCompositeColumnCondition(tablestore.LO_AND)
	condition.AddFilter(tablestore.NewSingleColumnCondition("follower", tablestore.CT_EQUAL, follower))
	condition.AddFilter(tablestore.NewSingleColumnCondition("followee", tablestore.CT_EQUAL, followee))
	req := new(tablestore.UpdateRowChange)
	req.TableName = FollowRelationTableName
	req.SetCondition(tablestore.RowExistenceExpectation_EXPECT_EXIST)
	req.SetColumnCondition(condition)
	req.PutColumn("status", int64(status))
	_, err := t.client.UpdateRow(&tablestore.UpdateRowRequest{
		UpdateRowChange: req,
	})
	return err
}

func (t *TableStoreFollowRelationDAO) CountFollower(ctx context.Context, uid int64) (int64, error) {
	req := &tablestore.SQLQueryRequest{Query: fmt.Sprintf("SELECT COUNT(follower) AS cnt FROM %s WHERE followee=%d AND status=%d", FollowRelationTableName, uid, FollowRelationStatusActive)}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return 0, err
	}
	resultSet := resp.ResultSet
	if resultSet.HasNext() {
		row := resultSet.Next()
		return row.GetInt64ByName("cnt")
	}
	return 0, ErrFollowNotFound
}

func (t *TableStoreFollowRelationDAO) CountFollowee(ctx context.Context, uid int64) (int64, error) {
	req := &tablestore.SQLQueryRequest{Query: fmt.Sprintf("SELECT COUNT(followee) AS cnt FROM %s WHERE followee=%d AND status=%d", FollowRelationTableName, uid, FollowRelationStatusActive)}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return 0, err
	}
	resultSet := resp.ResultSet
	if resultSet.HasNext() {
		row := resultSet.Next()
		return row.GetInt64ByName("cnt")
	}
	return 0, ErrFollowNotFound
}

func (t *TableStoreFollowRelationDAO) rowToEntity(row tablestore.SQLRow) FollowRelation {
	var res FollowRelation
	res.Id, _ = row.GetInt64ByName("id")
	res.Follower, _ = row.GetInt64ByName("follower")
	res.Followee, _ = row.GetInt64ByName("followee")
	status, _ := row.GetInt64ByName("status")
	res.Status = uint8(status)
	res.CreateTime, _ = row.GetInt64ByName("create_time")
	res.UpdateTime, _ = row.GetInt64ByName("update_time")
	return res
}
