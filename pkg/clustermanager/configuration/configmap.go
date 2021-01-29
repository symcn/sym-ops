package configuration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/symcn/sym-ops/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	listConfigmapTimeout = time.Second * 5
)

// cfgWithConfigmap clusterconfiguration manager with kubernetes configmap
type cfgWithConfigmap struct {
	kubeInterface kubernetes.Interface
	namespace     string
	label         map[string]string
	dataKey       string
	statusKey     string
	timeout       time.Duration
}

// NewClusterCfgManagerWithCM build cfgWithConfigmap
func NewClusterCfgManagerWithCM(kubeInterface kubernetes.Interface, namespace string, label map[string]string, dataKey, statusKey string) types.ClusterConfigurationManager {
	return &cfgWithConfigmap{
		kubeInterface: kubeInterface,
		namespace:     namespace,
		label:         label,
		dataKey:       dataKey,
		statusKey:     statusKey,
	}
}

func (cc *cfgWithConfigmap) GetAll() ([]types.ClusterCfgInfo, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), listConfigmapTimeout)
	defer cancel()

	labelSelectors := make([]string, 0, len(cc.label))
	for k, v := range cc.label {
		if k != "" && v != "" {
			labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", k, v))
		}
	}

	cmlist, err := cc.kubeInterface.CoreV1().ConfigMaps(cc.namespace).List(ctx, metav1.ListOptions{LabelSelector: strings.Join(labelSelectors, ",")})
	if err != nil {
		return nil, fmt.Errorf("Get clusterconfiguration with configmap failed namespace:%s label:%+v err:%+v", cc.namespace, cc.label, err)
	}
	return Configmap2ClusterCfgInfo(cmlist, cc.dataKey, cc.statusKey), nil
}

// Configmap2ClusterCfgInfo configmaplist to clusterconfiguration info
func Configmap2ClusterCfgInfo(cmlist *v1.ConfigMapList, dataKey, statusKey string) []types.ClusterCfgInfo {
	list := make([]types.ClusterCfgInfo, 0, len(cmlist.Items))

	for _, cm := range cmlist.Items {
		kubecfg, ok := cm.Data[dataKey]
		if !ok {
			// if not exist dataKey continue
			continue
		}
		if status, ok := cm.Data[statusKey]; ok && !strings.EqualFold(status, "true") {
			// if status not exist means should connected
			// status is equal true means should connected
			// otherwise disconnected
			continue
		}
		list = append(list, BuildClusterCfgInfo(cm.Name, types.KubeConfigTypeRawString, kubecfg, ""))
	}

	return list
}
