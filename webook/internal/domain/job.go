package domain

type Job struct {
	Id         int64
	CancelFunc func()
}
