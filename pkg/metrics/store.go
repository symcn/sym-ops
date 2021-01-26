package metrics

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/symcn/sym-ops/pkg/types"
)

// MaxLabelCount max label count limit
const MaxLabelCount = 20

var (
	defaultEndpoint = "/metrics"
	defaultStore    *store
	// ErrLabelCountExceeded error label count exceeded
	ErrLabelCountExceeded = fmt.Errorf("label count exceeded, max is %d", MaxLabelCount)
)

type store struct {
	l       sync.RWMutex
	metrics map[string]types.Metrics
}

type metrics struct {
	typ     string
	prefix  string
	buckets []float64
	col     []prometheus.Collector
}

func init() {
	defaultStore = &store{
		metrics: make(map[string]types.Metrics, 100),
	}
}

func (m *metrics) Counter(key string) prometheus.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: flattenKey(m.prefix + key),
	})
	m.registerPrometheus(counter)
	return counter
}

func (m *metrics) Gauge(key string) prometheus.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: flattenKey(m.prefix + key),
	})
	m.registerPrometheus(gauge)
	return gauge
}

func (m *metrics) Histogram(key string) prometheus.Histogram {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    flattenKey(m.prefix + key),
		Buckets: m.buckets,
	})
	m.registerPrometheus(histogram)
	return histogram
}

func (m *metrics) UnregisterAll() {
	for _, col := range m.col {
		prometheus.Unregister(col)
	}
}

func (m *metrics) registerPrometheus(c prometheus.Collector) {
	prometheus.MustRegister(c)
	m.col = append(m.col, c)
}

// Only [a-zA-Z0-9:_] are valid in metric names, any other characters should be sanitized to an underscore.
var flattenRegexp = regexp.MustCompile("[^a-zA-Z0-9_:]")

func flattenKey(key string) string {
	return flattenRegexp.ReplaceAllString(key, "_")
}

func resetAll() {
	defaultStore.l.Lock()
	defer defaultStore.l.Unlock()

	for _, m := range defaultStore.metrics {
		m.UnregisterAll()
	}
	defaultStore.metrics = make(map[string]types.Metrics, 100)
}

func sortedLabels(labels map[string]string) (keys, values []string) {
	keys = make([]string, 0, len(labels))
	values = make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		values = append(values, labels[k])
	}
	return
}

func fullName(typ string, labels map[string]string) (fullName string) {
	keys, values := sortedLabels(labels)

	pair := make([]string, 0, len(keys))
	for i := 0; i < len(keys); i++ {
		pair = append(pair, keys[i]+"."+values[i])
	}
	fullName = typ + "." + strings.Join(pair, ".")
	return
}
