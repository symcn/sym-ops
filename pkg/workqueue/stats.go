package workqueue

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/symcn/sym-ops/pkg/metrics"
)

var (
	metricTypePre      = "workqueue_"
	workqueueLabelName = "name"
)

// metrics key with labels
const (
	DequeueTotal          = "dequeue_total"
	UnExpectedObjTotal    = "unexpected_obj_total"
	ReconcileSuccTotal    = "reconcile_succ_total"
	ReconcileFailTotal    = "reconcile_fail_total"
	ReconcileTimeDuration = "reconcile_duration"
	RequeueAfterTotal     = "requeue_after_total"
	RequeueRateLimitTotal = "requeue_rate_limit_total"
)

type stats struct {
	Dequeue           prometheus.Counter
	UnExpectedObj     prometheus.Counter
	ReconcileSucc     prometheus.Counter
	ReconcileFail     prometheus.Counter
	ReconcileDuration prometheus.Histogram
	RequeueAfter      prometheus.Counter
	RequeueRateLimit  prometheus.Counter
}

func buildStats(name string) (*stats, error) {
	metric, err := metrics.NewMetrics(metricTypePre+name, nil, nil)
	if err != nil {
		return nil, err
	}

	return &stats{
		Dequeue:           metric.Counter(DequeueTotal),
		UnExpectedObj:     metric.Counter(UnExpectedObjTotal),
		ReconcileSucc:     metric.Counter(ReconcileSuccTotal),
		ReconcileFail:     metric.Counter(ReconcileFailTotal),
		ReconcileDuration: metric.Histogram(ReconcileTimeDuration),
		RequeueAfter:      metric.Counter(RequeueAfterTotal),
		RequeueRateLimit:  metric.Counter(RequeueRateLimitTotal),
	}, nil
}
