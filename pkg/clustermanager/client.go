package clustermanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	controllers "sigs.k8s.io/controller-runtime"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	rtmanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

type client struct {
	*ClientConfig

	stopCh    chan struct{}
	connected bool
	started   bool

	KubeRestConfig *rest.Config
	KubeInterface  kubernetes.Interface

	CtrlRtManager rtmanager.Manager
	CtrlRtCache   rtcache.Cache
	CtrlRtClient  rtclient.Client
}

// NewMingleClient build types.MingleClient
func NewMingleClient(cfg *ClientConfig) (types.MingleClient, error) {
	cli := &client{
		ClientConfig: cfg,
		stopCh:       make(chan struct{}, 0),
	}

	// 1. pre check
	if err := cli.preCheck(); err != nil {
		return nil, err
	}

	// 2. initialization
	if err := cli.initialization(); err != nil {
		return nil, err
	}

	return cli, nil
}

func (c *client) preCheck() error {
	if c.ClientConfig == nil {
		return errors.New("client configuration is empty")
	}

	// clusterconfig and cluster name must not empty
	if c.ClusterCfg == nil || c.ClusterCfg.GetName() == "" {
		return errors.New("cluster info is empty or cluster name is empty")
	}

	// cluster scheme must not empty
	if c.Scheme == nil {
		return errors.New("scheme is empty")
	}

	return nil
}

func (c *client) initialization() error {
	var err error
	// Step 1. build restconfig
	c.KubeRestConfig, err = buildClientCmd(c.ClusterCfg, c.SetKubeRestConfigFnList)
	if err != nil {
		return fmt.Errorf("cluster %s build kubernetes failed %+v", c.ClusterCfg.GetName(), err)
	}

	// Step 2. build kubernetes interface
	c.KubeInterface, err = buildKubeInterface(c.KubeRestConfig)
	if err != nil {
		return fmt.Errorf("cluster %s build kubernetes interface failed %+v", c.ClusterCfg.GetName(), err)
	}

	// Step 3. build controller-runtime manager
	c.CtrlRtManager, err = controllers.NewManager(c.KubeRestConfig, rtmanager.Options{
		Scheme:                  c.Scheme,
		SyncPeriod:              &c.SyncPeriod,
		LeaderElection:          c.LeaderElection,
		LeaderElectionNamespace: c.LeaderElectionNamespace,
		LeaderElectionID:        c.LeaderElectionID,
		MetricsBindAddress:      "0",
		HealthProbeBindAddress:  "0",
	})
	if err != nil {
		return fmt.Errorf("cluster %s build controller-runtime manager failed %+v", c.ClusterCfg.GetName(), err)
	}
	c.CtrlRtClient = c.CtrlRtManager.GetClient()
	c.CtrlRtCache = c.CtrlRtManager.GetCache()

	return nil
}

func (c *client) autoHealthCheck() {
	handler := func() {
		ok, err := healthRequestWithTimeout(c.KubeInterface, c.ExecTimeout)
		if err != nil {
			klog.Errorf("cluster %s check healthy failed %+v", c.ClusterCfg.GetName(), err)
		}
		c.connected = ok
	}

	// first check
	handler()

	// it will pointless when interval less than 1s
	if c.HealthCheckInterval < time.Second {
		klog.Warningf("cluster %s not enabled healthy check, interval must be greater than 1s")
		return
	}

	timer := time.NewTicker(c.HealthCheckInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			handler()
		case <-c.stopCh:
			return
		}
	}
}

// Start client and blocks until the context is cancelled
// Returns an error if there is an error starting
func (c *client) Start(ctx context.Context) error {
	if c.started {
		return fmt.Errorf("client %s can't repeat start", c.ClusterCfg.GetName())
	}
	c.started = true

	var err error

	go func() {
		err = c.CtrlRtManager.Start(ctx)
		if err != nil {
			klog.Errorf("start cluster %s error %+v", c.ClusterCfg.GetName(), err)
			close(c.stopCh)
		}
	}()

	// health check
	go c.autoHealthCheck()

	select {
	case <-ctx.Done():
		return err
	case <-c.stopCh:
		return err
	}
}

// IsConnected return connected status
func (c *client) IsConnected() bool {
	return c.connected
}