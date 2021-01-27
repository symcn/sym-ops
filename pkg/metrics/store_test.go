package metrics

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_metrics_Counter(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		m    *metrics
		args args
		want prometheus.Counter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Counter(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metrics.Counter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metrics_Gauge(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		m    *metrics
		args args
		want prometheus.Gauge
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Gauge(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metrics.Gauge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metrics_Histogram(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		m    *metrics
		args args
		want prometheus.Histogram
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Histogram(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metrics.Histogram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metrics_UnregisterAll(t *testing.T) {
	tests := []struct {
		name string
		m    *metrics
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.UnregisterAll()
		})
	}
}

func Test_registerPrometheus(t *testing.T) {
	defer resetAll()

	m := buildMetrics()

	type args struct {
		c prometheus.Collector
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case 1",
			args: args{
				c: prometheus.NewCounter(prometheus.CounterOpts{Name: "case1"}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.registerPrometheus(tt.args.c)
		})
	}
}

func Test_flattenKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				key: "aa_-:,aa",
			},
			want: "aa__:_aa",
		},
		{
			name: "case 2",
			args: args{
				key: " ",
			},
			want: "_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenKey(tt.args.key); got != tt.want {
				t.Errorf("flattenKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resetAll(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "case 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAll()
		})
	}
}

func Test_sortedLabels(t *testing.T) {
	type args struct {
		labels map[string]string
	}
	tests := []struct {
		name       string
		args       args
		wantKeys   []string
		wantValues []string
	}{
		{
			name: "case 1",
			args: args{
				labels: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			wantKeys:   []string{"k1", "k2"},
			wantValues: []string{"v1", "v2"},
		},
		{
			name: "case 2",
			args: args{
				labels: map[string]string{
					"k1": "v2",
					"k2": "v1",
				},
			},
			wantKeys:   []string{"k1", "k2"},
			wantValues: []string{"v2", "v1"},
		},
		{
			name: "case 3",
			args: args{
				labels: map[string]string{
					"":   "v2",
					"k2": "",
				},
			},
			wantKeys:   []string{"", "k2"},
			wantValues: []string{"v2", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeys, gotValues := sortedLabels(tt.args.labels)
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("sortedLabels() gotKeys = %v, want %v", gotKeys, tt.wantKeys)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("sortedLabels() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}
}

func Test_fullName(t *testing.T) {
	type args struct {
		typ    string
		labels map[string]string
	}
	tests := []struct {
		name         string
		args         args
		wantFullName string
	}{
		{
			name: "case 1",
			args: args{
				typ: "reconcile",
				labels: map[string]string{
					"k1": "v1",
				},
			},
			wantFullName: "reconcile.k1.v1",
		},
		{
			name: "case 2",
			args: args{
				typ: "reconcile",
				labels: map[string]string{
					"k1": "",
				},
			},
			wantFullName: "reconcile.k1.",
		},
		{
			name: "case 3",
			args: args{
				typ:    "reconcile",
				labels: nil,
			},
			wantFullName: "reconcile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFullName := fullName(tt.args.typ, tt.args.labels); gotFullName != tt.wantFullName {
				t.Errorf("fullName() = %v, want %v", gotFullName, tt.wantFullName)
			}
		})
	}
}

func buildMetrics() *metrics {
	return &metrics{
		typ: "sym-ops",
		col: []prometheus.Collector{},
	}
}
