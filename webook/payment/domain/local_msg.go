package domain

import "time"

type Msg struct {
	Id         int64
	Content    string
	CreateTime time.Time
}
