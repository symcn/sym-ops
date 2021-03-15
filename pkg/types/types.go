package types

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Scheme build runtime.scheme
var (
	Scheme = runtime.NewScheme()
)

// resource kind
const (
	ServiceKind       = "Service"
	DeploymentKind    = "Deployment"
	StatefulSetKind   = "StatefulSet"
	JobKind           = "Job"
	AppsetKind        = "Appset"
	AdvdeploymentKind = "Advdeployment"
)

// Metrics enum
var (
	MetricAverageUtilization = "AverageUtilization"
	MetricAverageValue       = "AverageValue"
)

// VersionSep version separation
var (
	VersionSep = "/"
)
