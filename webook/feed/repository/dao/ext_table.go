package dao

type FeedEvent struct {
	Id   int64
	Type string
}

type ArticleEvent struct {
	Id  int64
	Fid int64 // 指向 FeedEvent
	Aid int64 // 指向 Article
}
