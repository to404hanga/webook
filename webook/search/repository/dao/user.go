package dao

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/olivere/elastic/v7"
)

const UserIndexName = "user_index"

type UserElasticSearchDAO struct {
	client *elastic.Client
}

var _ UserDAO = (*UserElasticSearchDAO)(nil)

func NewUserElasticSearchDAO(client *elastic.Client) UserDAO {
	return &UserElasticSearchDAO{client: client}
}

func (u *UserElasticSearchDAO) InputUser(ctx context.Context, user User) error {
	_, err := u.client.Index().Index(UserIndexName).Id(strconv.FormatInt(user.Id, 10)).BodyJson(user).Do(ctx)
	return err
}

func (u *UserElasticSearchDAO) Search(ctx context.Context, keywords []string) ([]User, error) {
	queryString := strings.Join(keywords, " ")
	query := elastic.NewMatchQuery("nickname", queryString)
	resp, err := u.client.Search(UserIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var user User
		err = json.Unmarshal(hit.Source, &user)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return res, nil
}
