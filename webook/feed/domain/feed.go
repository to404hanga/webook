package domain

import "time"

// 以 A 发表了一篇文章为例
//
// * 如果是 Pull Event，拉模型，那么 Uid 是 A 的 Id
//
// * 如果是 Push Event，推模型，那么 Uid 是 A 的某个粉丝的 Id
type FeedEvent struct {
	Id         int64
	Uid        int64
	Type       string
	CreateTime time.Time
	Ext        ExtendFields
}
