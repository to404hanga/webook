package domain

type Amount struct {
	Currency string // 该字段用于支持国际化
	Total    int64
}

type Payment struct {
	Amt         Amount
	BizTradeNO  string
	Description string
	Status      PaymentStatus
	TxnID       string
}

type PaymentStatus uint8

func (s PaymentStatus) AsUint8() uint8 {
	return uint8(s)
}

const (
	PaymentStatusUnknown PaymentStatus = iota
	PaymentStatusInit
	PaymentStatusSuccess
	PaymentStatusFailed
	PaymentStatusRefund // 退款
)
