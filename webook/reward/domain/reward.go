package domain

type Target struct {
	Biz     string
	BizId   int64
	BizName string
	Uid     int64 // 打赏的目标用户
}

type Reward struct {
	Id     int64
	Uid    int64
	Target Target
	Amt    int64
	Status RewardStatus
}

// Completed 判断是否已经完成，即是否处理了支付回调
func (r Reward) Completed() bool {
	return r.Status == RewardStatusFailed || r.Status == RewardStatusPayed
}

type RewardStatus uint8

func (r RewardStatus) AsUint8() uint8 {
	return uint8(r)
}

const (
	RewardStatusUnknown RewardStatus = iota
	RewardStatusInit
	RewardStatusPayed
	RewardStatusFailed
)

type CodeURL struct {
	Rid int64
	URL string
}
