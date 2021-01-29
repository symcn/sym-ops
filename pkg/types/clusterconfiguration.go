package types

// KubeConfigType kubeconfig type
type KubeConfigType string

// KubeConfigTypeFile file
// KubeConfigTypeRawString rawstring
const (
	KubeConfigTypeFile      KubeConfigType = "file"
	KubeConfigTypeRawString KubeConfigType = "rawstring"
)

// ClusterConfigurationManager clusterconfiguration manager
type ClusterConfigurationManager interface {
	// GetAll return all need connect kubernetes cluster
	// return clusterconfiguration slice and error
	// if first get 3 clusterinfo, such as a, b, c
	// second get 2 clusterinfo, such as a, b, will means have one cluster (c) should be shutdown
	GetAll() ([]ClusterCfgInfo, error)
}

// ClusterCfgInfo clusterconfiguration info
type ClusterCfgInfo interface {
	// GetName return cluster Name
	GetName() string

	// GetKubeConfigType return kubeconfig type
	GetKubeConfigType() KubeConfigType

	// GetKubeConfig return kubeconfig such as file path or rawstring, reference KubeConfigType
	GetKubeConfig() string

	// GetKubeContext return kubeconfig context
	// if kubeconfig have multi cluster info, use context choice one cluster connect
	GetKubeContext() string
}
