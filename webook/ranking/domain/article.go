package domain

import "time"

type Article struct {
	Id         int64
	Title      string
	Content    string
	Author     Author
	Status     ArticleStatus
	CreateTime time.Time
	UpdateTime time.Time
}

func (a Article) Abstract() string {
	str := []rune(a.Content)
	if len(str) > 128 {
		return string(str[:128]) + "..."
	}
	return string(str)
}

type ArticleStatus uint8

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

type Author struct {
	Id   int64
	Name string
}
