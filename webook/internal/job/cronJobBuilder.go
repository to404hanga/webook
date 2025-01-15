package job

import (
	"strconv"
	"time"
	"webook/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
)

type CronJobBuilder struct {
	l      logger.Logger
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(l logger.Logger, opt prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opt, []string{"job", "success"})
	return &CronJobBuilder{
		l:      l,
		vector: vector,
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		start := time.Now()
		b.l.Debug("Job 开始运行", logger.String("name", name))
		err := job.Run()
		if err != nil {
			b.l.Error("Job 执行失败", logger.String("name", name), logger.Error(err))
		}
		b.l.Debug("Job 运行完毕", logger.String("name", name))
		duration := time.Since(start).Milliseconds()
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration))
	})
}

type cronJobAdapterFunc func()

func (f cronJobAdapterFunc) Run() {
	f()
}
