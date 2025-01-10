package web

type (
	ArticleVo struct {
		Id         int64  `json:"id,omitempty"`
		Title      string `json:"title,omitempty"`
		Abstract   string `json:"abstract,omitempty"`
		Content    string `json:"content,omitempty"`
		AuthorId   int64  `json:"authorId,omitempty"`
		AuthorName string `json:"authorName,omitempty"`
		Status     uint8  `json:"status,omitempty"`
		CreateTime string `json:"createTime,omitempty"`
		UpdateTime string `json:"updateTime,omitempty"`

		ReadCnt    int64 `json:"readCnt"`
		LikeCnt    int64 `json:"likeCnt"`
		CollectCnt int64 `json:"collectCnt"`
		Liked      bool  `json:"liked"`
		Collected  bool  `json:"collected"`
	}

	PublishReq struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	ArticleEditReq struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	ArticleWithdrawReq struct {
		Id int64
	}

	ArticleLikeReq struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}

	ArticleCollectReq struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}

	ArticleRewardReq struct {
		Id  int64 `json:"id"`
		Amt int64 `json:"amt"`
	}
)
