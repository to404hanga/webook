package prometheus

import (
	"context"
	"time"
	"webook/sms/service"

	"github.com/prometheus/client_golang/prometheus"
)

type Decorator struct {
	svc    service.Service
	vector *prometheus.SummaryVec
}

var _ service.Service = (*Decorator)(nil)

func NewDecorator(svc service.Service, opt prometheus.SummaryOpts) service.Service {
	return &Decorator{
		svc:    svc,
		vector: prometheus.NewSummaryVec(opt, []string{"tpl_id"}),
	}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		d.vector.WithLabelValues(tplId).Observe(float64(duration))
	}()
	return d.svc.Send(ctx, tplId, args, numbers...)
}
