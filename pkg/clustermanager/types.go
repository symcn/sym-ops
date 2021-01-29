package clustermanager

import (
	"time"

	"github.com/symcn/sym-ops/pkg/clustermanager/configuration"
	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var (
	defaultSyncPeriod          = time.Minute * 30
	defaultHealthCheckInterval = time.Second * 5
	defaultExecTimeout         = time.Second * 5
	defaultClusterName         = "meta"
	defaultQPS                 = 100
	defaultBurst               = 120
)

type option struct {
	scheme                  *runtime.Scheme
	metricsBindAddress      int
	leaderElection          bool
	leaderElectionNamespace string
	leaderElectionID        string
	resyncPeriod            time.Duration

	clusterCfgManager types.ClusterConfigurationManager
	setKubeRestConfig types.SetKubeRestConfig
}

// ClientConfig client configuration
type ClientConfig struct {
	Scheme                  *runtime.Scheme
	LeaderElection          bool
	LeaderElectionNamespace string
	LeaderElectionID        string
	SyncPeriod              time.Duration

	HealthCheckInterval     time.Duration
	ExecTimeout             time.Duration
	ClusterCfg              types.ClusterCfgInfo
	SetKubeRestConfigFnList []types.SetKubeRestConfig
}

// SingleClientConfig use default config
// if scheme is empty use default Kubernetes resource
// disable leader
// kubeconfig use default ~/.kube/config or Kubernetes cluster internal config
func SingleClientConfig(scheme *runtime.Scheme) *ClientConfig {
	if scheme == nil {
		scheme = runtime.NewScheme()
		clientgoscheme.AddToScheme(scheme)
	}

	return &ClientConfig{
		Scheme:              scheme,
		LeaderElection:      false,
		SyncPeriod:          defaultSyncPeriod,
		HealthCheckInterval: defaultHealthCheckInterval,
		ExecTimeout:         defaultExecTimeout,
		ClusterCfg:          configuration.BuildDefaultClusterCfgInfo(defaultClusterName),
		SetKubeRestConfigFnList: []types.SetKubeRestConfig{
			func(config *rest.Config) {
				config.QPS = float32(defaultQPS)
				config.Burst = defaultBurst
			},
		},
	}
}
