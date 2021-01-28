package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/symcn/sym-ops/pkg/types"
)

// NewMetrics returns a metrics
func NewMetrics(typ string, labels map[string]string) (types.Metrics, error) {
	if len(labels) > MaxLabelCount {
		return nil, ErrLabelCountExceeded
	}

	defaultStore.l.Lock()
	defer defaultStore.l.Unlock()

	if col, ok := defaultStore.metrics[typ]; ok {
		return col, nil
	}
	stats := &metrics{
		typ:    typ,
		prefix: fullName(typ, labels) + ".",
		col:    []prometheus.Collector{},
	}
	defaultStore.metrics[typ] = stats
	return stats, nil
}

// RegisterHTTPHandler register metrics with http mode
func RegisterHTTPHandler(f func(pattern string, handler http.Handler)) {
	f(defaultEndpoint, promhttp.Handler())
}
