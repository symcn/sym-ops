package clustermanager

import (
	"errors"
	"fmt"

	"github.com/symcn/sym-ops/pkg/clustermanager/handler"
	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/client-go/tools/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AddResourceEventHandler loop each mingleclient invoke AddResourceEventHandler
func (mc *multiclient) AddResourceEventHandler(obj rtclient.Object, handler cache.ResourceEventHandler) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for _, cli := range mc.mingleClientMap {
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
	for _, cli := range mc.mingleClientMap {
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
	for _, cli := range mc.mingleClientMap {
		err = cli.SetIndexField(obj, field, extractValue)
		if err != nil {
			return fmt.Errorf("cluster %s SetIndexField failed %+v", cli.GetClusterCfgInfo().GetName(), err)
		}
	}
	return nil
}

// Watch takes events provided by a Source and uses the EventHandler to
// enqueue reconcile.Requests in response to the events.
//
// Watch may be provided one or more Predicates to filter events before
// they are given to the EventHandler.  Events will be passed to the
// EventHandler if all provided Predicates evaluate to true.
func (mc *multiclient) Watch(obj rtclient.Object, queue types.WorkQueue, evtHandler types.EventHandler, predicates ...types.Predicate) error {
	if queue == nil {
		return errors.New("types.WorkQueue is nil")
	}
	err := mc.AddResourceEventHandler(obj, handler.NewResourceEventHandler(queue, evtHandler, predicates...))
	if err != nil {
		return fmt.Errorf("Watch resource failed %+v", err)
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

	for _, cli := range mc.mingleClientMap {
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

	if cli, ok := mc.mingleClientMap[name]; ok {
		return cli, nil
	}
	return nil, fmt.Errorf(ErrClientNotExist, name)
}

// GetConnectedWithName returns MingleClient object with name and status is connected
func (mc *multiclient) GetConnectedWithName(name string) (types.MingleClient, error) {
	mc.l.Lock()
	defer mc.l.Unlock()

	if cli, ok := mc.mingleClientMap[name]; ok {
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

	list := make([]types.MingleClient, 0, len(mc.mingleClientMap))
	for _, cli := range mc.mingleClientMap {
		list = append(list, cli)
	}
	return list
}

// GetAllConnected returns all MingleClient which status is connected
func (mc *multiclient) GetAllConnected() []types.MingleClient {
	mc.l.Lock()
	defer mc.l.Unlock()

	list := make([]types.MingleClient, 0, len(mc.mingleClientMap))
	for _, cli := range mc.mingleClientMap {
		if cli.IsConnected() {
			list = append(list, cli)
		}
	}
	return list
}

// RegistryBeforAfterHandler registry BeforeStartHandle
func (mc *multiclient) RegistryBeforAfterHandler(handler types.BeforeStartHandle) {
	mc.beforStartHandleList = append(mc.beforStartHandleList, handler)
}
