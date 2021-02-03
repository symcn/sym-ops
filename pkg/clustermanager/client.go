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
	*Options
	*ClientOptions

	stopCh         chan struct{}
	connected      bool
	started        bool
	internalCancel context.CancelFunc
	informerList   []rtcache.Informer

	kubeRestConfig *rest.Config
	kubeInterface  kubernetes.Interface

	ctrlRtManager rtmanager.Manager
	ctrlRtCache   rtcache.Cache
	ctrlRtClient  rtclient.Client
}

// NewMingleClient build types.MingleClient
func NewMingleClient(clinetOpt *ClientOptions, opt *Options) (types.MingleClient, error) {
	cli := &client{
		Options:       opt,
		ClientOptions: clinetOpt,
		stopCh:        make(chan struct{}, 0),
		informerList:  []rtcache.Informer{},
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
	if c.ClientOptions == nil {
		return errors.New("client configuration is empty")
	}

	// clusterconfig and cluster name must not empty
	if c.ClientOptions.ClusterCfg == nil || c.ClientOptions.ClusterCfg.GetName() == "" {
		return errors.New("cluster info is empty or cluster name is empty")
	}

	// cluster scheme must not empty
	if c.Options.Scheme == nil {
		return errors.New("scheme is empty")
	}

	if c.Options.ExecTimeout < minExectimeout {
		klog.Warningf("exectimeout should lager than 100ms, too small will return timeout mostly, use default %v", defaultExecTimeout)
		c.Options.ExecTimeout = defaultExecTimeout
	}

	return nil
}

func (c *client) initialization() error {
	var err error
	// Step 1. build restconfig
	c.kubeRestConfig, err = buildClientCmd(c.ClusterCfg, c.SetKubeRestConfigFnList)
	if err != nil {
		return fmt.Errorf("cluster %s build kubernetes failed %+v", c.ClusterCfg.GetName(), err)
	}

	// Step 2. build kubernetes interface
	c.kubeInterface, err = buildKubeInterface(c.kubeRestConfig)
	if err != nil {
		return fmt.Errorf("cluster %s build kubernetes interface failed %+v", c.ClusterCfg.GetName(), err)
	}

	// Step 3. build controller-runtime manager
	c.ctrlRtManager, err = controllers.NewManager(c.kubeRestConfig, rtmanager.Options{
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
	c.ctrlRtClient = c.ctrlRtManager.GetClient()
	c.ctrlRtCache = c.ctrlRtManager.GetCache()

	return nil
}

func (c *client) autoHealthCheck() {
	handler := func() {
		ok, err := healthRequestWithTimeout(c.kubeInterface.Discovery().RESTClient(), c.ExecTimeout)
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

	timer := time.NewTicker(c.Options.HealthCheckInterval)
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
	ctx, c.internalCancel = context.WithCancel(ctx)

	go func() {
		err = c.ctrlRtManager.Start(ctx)
		if err != nil {
			klog.Errorf("start cluster %s error %+v", c.ClusterCfg.GetName(), err)
			close(c.stopCh)
		}
		klog.Warningf("cluster %s stoped.", c.ClusterCfg.GetName())
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

// Stop stop mingle client, just use with multiclient, not recommend use direct
func (c *client) Stop() {
	if c.internalCancel == nil {
		return
	}
	c.internalCancel()
}

// IsConnected return connected status
func (c *client) IsConnected() bool {
	return c.connected
}

// GetClusterCfgInfo returns cluster configuration info
func (c *client) GetClusterCfgInfo() types.ClusterCfgInfo {
	return c.ClientOptions.ClusterCfg
}
