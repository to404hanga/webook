package domain

import "time"

type Comment struct {
	Id            int64     `json:"id"`
	Commentator   User      `json:"user"`
	Biz           string    `json:"biz"`
	BizId         int64     `json:"bizId"`
	Content       string    `json:"content"`
	RootComment   *Comment  `json:"rootComment"`
	ParentComment *Comment  `json:"parentComment"`
	Children      []Comment `json:"children"`
	CreateTime    time.Time `json:"createTime"`
	UpdateTime    time.Time `json:"updateTime"`
}

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
