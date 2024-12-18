package domain

import "time"

type User struct {
	Id         int64
	Email      string
	Password   string
	CreateTime time.Time // UTC 0 的时区
	Nickname   string
	Birthday   time.Time
	AboutMe    string
}
