package service

import (
	"context"
	"time"
	"webook/oauth2/domain"

	"github.com/prometheus/client_golang/prometheus"
)

type Decorator struct {
	Service
	sum prometheus.Summary
}

func NewDecorator(svc Service, namespace string, subsystem string, instanceId string, name string) *Decorator {
	sum := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      name,
		Namespace: namespace,
		Subsystem: subsystem,
		ConstLabels: map[string]string{
			"instance_id": instanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.95:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	prometheus.MustRegister(sum)
	return &Decorator{
		Service: svc,
		sum:     sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
