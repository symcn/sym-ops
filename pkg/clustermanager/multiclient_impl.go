package clustermanager

import (
	"fmt"

	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/client-go/tools/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AddResourceEventHandler loop each mingleclient invoke AddResourceEventHandler
func (mc *multiclient) AddResourceEventHandler(obj rtclient.Object, handler cache.ResourceEventHandler) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for _, cli := range mc.clusterClientMap {
		err = cli.AddResourceEventHandler(obj, handler)
		if err != nil {
			return fmt.Errorf("cluster %s AddResourceEventHandler failed %+v", cli.GetClusterCfgInfo().GetName(), err)
		}
	}
	return nil
}

// TriggerSync just trigger each mingleclient cache resource without handler
func (mc *multiclient) TriggerSync(obj rtclient.Object) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for _, cli := range mc.clusterClientMap {
		_, err = cli.GetInformer(obj)
		if err != nil {
			return fmt.Errorf("cluster %s TriggerSync failed %+v", cli.GetClusterCfgInfo().GetName(), err)
		}
	}
	return nil
}

// SetIndexField loop each mingleclient invoke SetIndexField
func (mc *multiclient) SetIndexField(obj rtclient.Object, field string, extractValue rtclient.IndexerFunc) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for _, cli := range mc.clusterClientMap {
		err = cli.SetIndexField(obj, field, extractValue)
		if err != nil {
			return fmt.Errorf("cluster %s SetIndexField failed %+v", cli.GetClusterCfgInfo().GetName(), err)
		}
	}
	return nil
}

// HasSynced return true if all mingleclient and all informers underlying store has synced
// !import if informerlist is empty, will return true
func (mc *multiclient) HasSynced() bool {
	if !mc.started {
		return false
	}

	mc.l.Lock()
	defer mc.l.Unlock()

	for _, cli := range mc.clusterClientMap {
		if !cli.HasSynced() {
			return false
		}
	}
	return true
}

// GetWithName returns MingleClient object with name
func (mc *multiclient) GetWithName(name string) (types.MingleClient, error) {
	mc.l.Lock()
	defer mc.l.Unlock()

	if cli, ok := mc.clusterClientMap[name]; ok {
		return cli, nil
	}
	return nil, fmt.Errorf(ErrClientNotExist, name)
}

// GetConnectedWithName returns MingleClient object with name and status is connected
func (mc *multiclient) GetConnectedWithName(name string) (types.MingleClient, error) {
	mc.l.Lock()
	defer mc.l.Unlock()

	if cli, ok := mc.clusterClientMap[name]; ok {
		if cli.IsConnected() {
			return cli, nil
		}
		return nil, fmt.Errorf(ErrClientNotConnected, name)
	}
	return nil, fmt.Errorf(ErrClientNotExist, name)
}

// GetAll returns all MingleClient
func (mc *multiclient) GetAll() []types.MingleClient {
	mc.l.Lock()
	defer mc.l.Unlock()

	list := make([]types.MingleClient, 0, len(mc.clusterClientMap))
	for _, cli := range mc.clusterClientMap {
		list = append(list, cli)
	}
	return list
}

// GetAllConnected returns all MingleClient which status is connected
func (mc *multiclient) GetAllConnected() []types.MingleClient {
	mc.l.Lock()
	defer mc.l.Unlock()

	list := make([]types.MingleClient, 0, len(mc.clusterClientMap))
	for _, cli := range mc.clusterClientMap {
		if cli.IsConnected() {
			list = append(list, cli)
		}
	}
	return list
}
