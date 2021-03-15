package types

var (
	// CurrentClusterName connected internal cluster name
	CurrentClusterName = "current"

	// MasterQueueName master feature queue name
	MasterQueueName = "master"
	// WorkerQueueName worker featrue queue name
	WorkerQueueName = "worker"
)

// MultiCluster configuration manager
var (
	MultiClusterCfgConfigmapNamespace = "sym-admin"
	MultiClusterCfgConfigmapLabels    = map[string]string{
		"ClusterOwner": "sym-admin",
	}
	MultiClusterCfgConfigmapDataKey   = "kubeconfig.yaml"
	MultiClusterCfgConfigmapStatusKey = "status"
)

// Filter info
var (
	FilterNamespaceAppset        = []string{"*"}
	FilterNamespaceAdvdeployment = []string{"*"}
)

// Finalizers
var (
	FinalizersStr = "sym-admin-finalizers"
)

// labels
var (
	ObserveMustLabelClusterName = "sym-cluster-info"
	ObserveMustLabelAppName     = "app"

	ServiceNameSuffix = "-svc"

	LabelKeyZone = "sym-available-zone"
)

// annotation
var (
	AnnotationsHpa        = "hpa.autoscaling.dmall.com/Hpa"
	AnnotationsHpaMetrics = "hpa.autoscaling.dmall.com/Metrics"
)

// HorizontalAPIVersion horizontal api version
var (
	HorizontalAPIVersion = "apps/v1"
)
