package clusterconfiguration

import "github.com/symcn/sym-ops/pkg/types"

type clusterCfgInfo struct {
	name           string
	kubeConfigType types.KubeConfigType
	kubeConfig     string
	kubeContext    string
}

// BuildClusterCfgInfo build types.ClusterCfgInfo
func BuildClusterCfgInfo(name string, kubeConfigType types.KubeConfigType, kubeConfig string, kubeContext string) types.ClusterCfgInfo {
	return &clusterCfgInfo{
		name:           name,
		kubeConfigType: kubeConfigType,
		kubeConfig:     kubeConfig,
		kubeContext:    kubeContext,
	}
}

func (c *clusterCfgInfo) GetName() string {
	return c.name
}

func (c *clusterCfgInfo) GetKubeConfigType() types.KubeConfigType {
	return c.kubeConfigType
}

func (c *clusterCfgInfo) GetKubeConfig() string {
	return c.kubeConfig
}

func (c *clusterCfgInfo) GetKubeContext() string {
	return c.kubeContext
}
