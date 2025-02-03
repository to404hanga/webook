package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	Id         int64
	Name       string
	Executor   string
	Cfg        string
	CancelFunc func()
	Expression string // cron 表达式
	NextTime   time.Time
}

func (j CronJob) Next(t time.Time) time.Time {
	c := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s, _ := c.Parse(j.Expression)
	return s.Next(t)
}
