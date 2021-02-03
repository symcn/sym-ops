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

	minExectimeout = time.Millisecond * 100
)

var (
	// ErrClientNotExist client not exist error
	ErrClientNotExist = "cluster %s not exist"
	// ErrClientNotConnected client disconnected
	ErrClientNotConnected = "cluster %s disconnected"
)

// Options options
type Options struct {
	Scheme                  *runtime.Scheme
	LeaderElection          bool
	LeaderElectionNamespace string
	LeaderElectionID        string
	SyncPeriod              time.Duration
	HealthCheckInterval     time.Duration
	ExecTimeout             time.Duration
	SetKubeRestConfigFnList []types.SetKubeRestConfig
}

// ClientOptions client configuration
type ClientOptions struct {
	ClusterCfg types.ClusterCfgInfo
}

// MultiClientOptions multi client configuration
type MultiClientOptions struct {
	ClusterConfigurationManager types.ClusterConfigurationManager
	RebuildInterval             time.Duration
}

// SimpleClientOptions use default config
// kubeconfig use default ~/.kube/config or Kubernetes cluster internal config
func SimpleClientOptions() *ClientOptions {
	return &ClientOptions{
		ClusterCfg: configuration.BuildDefaultClusterCfgInfo(defaultClusterName),
	}
}

// DefaultOptions use default config
// if scheme is empty use default Kubernetes resource
// disable leader
func DefaultOptions(scheme *runtime.Scheme, qps, burst int) *Options {
	if scheme == nil {
		scheme = runtime.NewScheme()
		clientgoscheme.AddToScheme(scheme)
	}
	if qps < 1 {
		qps = defaultQPS
	}
	if burst < 1 {
		burst = defaultBurst
	}

	return &Options{
		Scheme:              scheme,
		LeaderElection:      false,
		SyncPeriod:          defaultSyncPeriod,
		HealthCheckInterval: defaultHealthCheckInterval,
		ExecTimeout:         defaultExecTimeout,
		SetKubeRestConfigFnList: []types.SetKubeRestConfig{
			func(config *rest.Config) {
				config.QPS = float32(qps)
				config.Burst = burst
			},
		},
	}
}
