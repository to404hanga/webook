package domain

type FollowRelation struct {
	Followee int64
	Follower int64
}

type FollowStatics struct {
	Followers int64
	Followees int64
}
