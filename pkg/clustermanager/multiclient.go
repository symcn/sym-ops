package clustermanager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

type multiclient struct {
	l                 sync.Mutex
	ctx               context.Context
	stopCh            chan struct{}
	started           bool
	rebuildInterval   time.Duration
	clusterClientMap  map[string]types.MingleClient
	clustercfgmanager types.ClusterConfigurationManager
	beforHandleList   []types.BeforeHandle

	scheme                  *runtime.Scheme
	leaderElection          bool
	leaderElectionNamespace string
	leaderElectionID        string
	syncPeriod              time.Duration

	healthCheckInterval     time.Duration
	execTimeout             time.Duration
	setKubeRestConfigFnList []types.SetKubeRestConfig
}

// NewMultiMingleClient build multiclient
func NewMultiMingleClient(rebuildInterval time.Duration, clustercfgmanager types.ClusterConfigurationManager, beforHandleList ...types.BeforeHandle) (types.MultiMingleClient, error) {
	return nil, nil
}

func (mc *multiclient) Start(ctx context.Context) error {
	if mc.started {
		return errors.New("multiclient can't repeat start")
	}
	mc.started = true

	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for _, cli := range mc.clusterClientMap {
		err = start(mc.ctx, cli, mc.beforHandleList...)
		if err != nil {
			return err
		}
	}

	select {
	case <-ctx.Done():
		close(mc.stopCh)
		return err
	}
}

func start(ctx context.Context, cli types.MingleClient, beforHandleList ...types.BeforeHandle) error {
	var err error
	for _, handler := range beforHandleList {
		err = handler(cli)
		if err != nil {
			return fmt.Errorf("invoke mingle client %s BeforeHandle failed %+v", cli.GetClusterCfgInfo().GetName(), err)
		}
	}

	go func() {
		err = cli.Start(ctx)
		if err != nil {
			klog.Error("start mingle client %s failed %+v", cli.GetClusterCfgInfo().GetName(), err)
		}
	}()

	return nil
}

func (mc *multiclient) autoRebuild() {
	if mc.rebuildInterval <= 0 {
		return
	}

	var err error
	timer := time.NewTicker(mc.rebuildInterval)
	for {
		select {
		case <-timer.C:
			err = mc.Rebuild()
			if err != nil {
				klog.Errorf("Rebuild failed %+v", err)
			}
		case <-mc.stopCh:
		}
	}
}

// Rebuild get clusterconfigurationmanager GetAll and rebuild clusterClientMap
func (mc *multiclient) Rebuild() error {
	if !mc.started {
		return nil
	}

	mc.l.Lock()
	defer mc.l.Unlock()

	newList, err := mc.clustercfgmanager.GetAll()
	if err != nil {
		return fmt.Errorf("get all cluster info failed %+v", err)
	}

	newCliMap := make(map[string]types.MingleClient, len(newList))
	// add and check new cluster
	for _, newClsCfg := range newList {
		// get old cluster info
		oldCli, exist := mc.clusterClientMap[newClsCfg.GetName()]
		if exist &&
			oldCli.GetClusterCfgInfo().GetKubeConfigType() == newClsCfg.GetKubeConfigType() &&
			oldCli.GetClusterCfgInfo().GetKubeConfig() == newClsCfg.GetKubeConfig() &&
			oldCli.GetClusterCfgInfo().GetKubeContext() == newClsCfg.GetKubeContext() {
			// kubetype, kubeconfig, kubecontext not modify
			newCliMap[oldCli.GetClusterCfgInfo().GetName()] = oldCli
			continue
		}

		// build new client
		// TODO complete logic
		cli, err := NewMingleClient(&ClientConfig{})
		if err != nil {
			klog.Error(err)
			continue
		}

		// start new client
		err = start(mc.ctx, cli, mc.beforHandleList...)
		if err != nil {
			klog.Error(err)
			continue
		}

		if exist {
			// kubeconfig modify, should stop old client
			oldCli.Stop()
		}

		newCliMap[newClsCfg.GetName()] = cli
		klog.Infof("auto add mingle client %s", newClsCfg.GetName())
	}

	// remove unexpect cluster
	for name, oldCli := range mc.clusterClientMap {
		if _, ok := newCliMap[name]; !ok {
			// not exist, should stop
			go func(cli types.MingleClient) {
				cli.Stop()
			}(oldCli)
		}
	}
	mc.clusterClientMap = newCliMap
	return nil
}
