package advdeployment

import (
	"k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
)

// AdvConfig extend config
type AdvConfig struct {
	RevisionHistoryLimit    int32
	ProgressDeadlineSeconds int32
	Debug                   bool

	MetricCPUValue *int32
	MetricMemValue *int32
}

var (
	defaultRevisionHistoryLimit    int32 = 10
	defaultProgressDeadlineSeconds int32 = 600
	defaultMetricCPUValue          int32 = 70
	defaultMetricMemValue          int32 = 70
)

// DefaultAdvConfig returns default AdvConfig
func DefaultAdvConfig() *AdvConfig {
	return &AdvConfig{
		RevisionHistoryLimit:    defaultRevisionHistoryLimit,
		ProgressDeadlineSeconds: defaultProgressDeadlineSeconds,
		Debug:                   false,
		MetricCPUValue:          &defaultMetricCPUValue,
		MetricMemValue:          &defaultMetricMemValue,
	}
}

func (w *worker) getDefaultMetric() []v2beta2.MetricSpec {
	return []v2beta2.MetricSpec{
		{
			Type: v2beta2.ResourceMetricSourceType,
			Resource: &v2beta2.ResourceMetricSource{
				Name: v1.ResourceCPU,
				Target: v2beta2.MetricTarget{
					Type:               v2beta2.UtilizationMetricType,
					AverageUtilization: w.conf.MetricCPUValue,
				},
			},
		},
		{
			Type: v2beta2.ResourceMetricSourceType,
			Resource: &v2beta2.ResourceMetricSource{
				Name: v1.ResourceMemory,
				Target: v2beta2.MetricTarget{
					Type:               v2beta2.UtilizationMetricType,
					AverageUtilization: w.conf.MetricMemValue,
				},
			},
		},
	}
}
