package types

import (
	"context"

	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	rtmanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

// SetKubeRestConfig set rest config
// such as QPS Burst
type SetKubeRestConfig func(config *rest.Config)

// BeforeStartHandle before Start exec this handle
// registry informer, when multi cluster manager add new cluster
// should record before handle, returns error will not start
type BeforeStartHandle func(cli MingleClient) error

// MingleClient mingle client
// wrap controller-runtime manager
type MingleClient interface {
	ResourceOperate

	// if dissatisfy can use this interface get Kubernetes resource
	KubernetesResource

	// if dissatisfy can use this interface get controller-runtime manager resource
	ControllerRuntimeManagerResource

	Controller

	// Start client and blocks until the context is cancelled
	// Returns an error if there is an error starting
	Start(ctx context.Context) error

	// Stop stop mingle client, just use with multiclient, not recommend use direct
	Stop()

	// IsConnected return connected status
	IsConnected() bool

	// GetClusterCfgInfo returns cluster configuration info
	GetClusterCfgInfo() ClusterCfgInfo
}

// ResourceOperate Kubernetes resource CRUD operate.
type ResourceOperate interface {
	// GetInformer fetches or constructs an informer for the given object that corresponds to a single
	// API kind and resource.
	GetInformer(obj rtclient.Object) (rtcache.Informer, error)

	// AddResourceEventHandler
	// 1. GetInformer
	// 2. Adds an event handler to the shared informer using the shared informer's resync
	//	period.  Events to a single handler are delivered sequentially, but there is no coordination
	//	between different handlers.
	AddResourceEventHandler(obj rtclient.Object, handler cache.ResourceEventHandler) error

	// IndexFields adds an index with the given field name on the given object type
	// by using the given function to extract the value for that field.  If you want
	// compatibility with the Kubernetes API server, only return one key, and only use
	// fields that the API server supports.  Otherwise, you can return multiple keys,
	// and "equality" in the field selector means that at least one key matches the value.
	// The FieldIndexer will automatically take care of indexing over namespace
	// and supporting efficient all-namespace queries.
	SetIndexField(obj rtclient.Object, field string, extractValue rtclient.IndexerFunc) error

	// HasSynced return true if all informers underlying store has synced
	// !import if informerlist is empty, will return true
	HasSynced() bool

	// Get retrieves an obj for the given object key from the Kubernetes Cluster with timeout.
	// obj must be a struct pointer so that obj can be updated with the response
	// returned by the Server.
	Get(key ktypes.NamespacedName, obj rtclient.Object) error

	// Create saves the object obj in the Kubernetes cluster with timeout.
	Create(obj rtclient.Object, opts ...rtclient.CreateOption) error

	// Delete deletes the given obj from Kubernetes cluster with timeout.
	Delete(obj rtclient.Object, opts ...rtclient.DeleteOption) error

	// Update updates the given obj in the Kubernetes cluster with timeout. obj must be a
	// struct pointer so that obj can be updated with the content returned by the Server.
	Update(obj rtclient.Object, opts ...rtclient.UpdateOption) error

	// Update updates the fields corresponding to the status subresource for the
	// given obj with timeout. obj must be a struct pointer so that obj can be updated
	// with the content returned by the Server.
	StatusUpdate(obj rtclient.Object, opts ...rtclient.UpdateOption) error

	// Patch patches the given obj in the Kubernetes cluster with timeout. obj must be a
	// struct pointer so that obj can be updated with the content returned by the Server.
	Patch(obj rtclient.Object, patch rtclient.Patch, opts ...rtclient.PatchOption) error

	// DeleteAllOf deletes all objects of the given type matching the given options with timeout.
	DeleteAllOf(obj rtclient.Object, opts ...rtclient.DeleteAllOfOption) error
}

// KubernetesResource Kubernetes object operate
type KubernetesResource interface {
	// GetRestConfig return Kubernetes rest Config
	GetKubeRestConfig() *rest.Config

	// GetKubeInterface return Kubernetes Interface.
	// kubernetes.ClientSet impl kubernetes.Interface
	GetKubeInterface() kubernetes.Interface
}

// ControllerRuntimeManagerResource controller-runtime manager resource
type ControllerRuntimeManagerResource interface {
	// GetCtrlRtManager return controller-runtime manager object
	GetCtrlRtManager() rtmanager.Manager

	// GetCtrlRtCache return controller-runtime cache object
	GetCtrlRtCache() rtcache.Cache

	// GetCtrlRtClient return controller-runtime client
	GetCtrlRtClient() rtclient.Client
}

// Controller implements a Kubernetes API. A Controller manages a work queue fed reconcile.Requests
// from source.Sources. Work is performed through the reconcile.Reconcile for each enqueued item.
// Work typically is reads and writes Kubernetes objectes to make the system state match the state specified
// in the object Spec.
type Controller interface {
	// Watch takes events provided by a Source and uses the EventHandler to
	// enqueue reconcile.Requests in response to the events.
	//
	// Watch may be provided one or more Predicates to filter events before
	// they are given to the EventHandler.  Events will be passed to the
	// EventHandler if all provided Predicates evaluate to true.
	Watch(src rtclient.Object, queue WorkQueue, handler EventHandler, predicates ...Predicate) error
}

// MultiMingleClient multi mingleclient
type MultiMingleClient interface {
	MultiMingleResource

	MultiClientOperate

	Controller

	// Rebuild get clusterconfigurationmanager GetAll and rebuild clusterClientMap
	Rebuild() error

	// HasSynced return true if all mingleclient and all informers underlying store has synced
	// !import if informerlist is empty, will return true
	HasSynced() bool

	// Start multiclient and blocks until the context is cancelled
	// Returns an error if there is an error starting
	Start(ctx context.Context) error
}

// MultiMingleResource multi MingleClient Resource
type MultiMingleResource interface {
	// AddResourceEventHandler loop each mingleclient invoke AddResourceEventHandler
	AddResourceEventHandler(obj rtclient.Object, handler cache.ResourceEventHandler) error

	// TriggerSync just trigger each mingleclient cache resource without handler
	TriggerSync(obj rtclient.Object) error

	// SetIndexField loop each mingleclient invoke SetIndexField
	SetIndexField(obj rtclient.Object, field string, extractValue rtclient.IndexerFunc) error
}

// MultiClientOperate multi client operate
type MultiClientOperate interface {
	// GetWithName returns MingleClient object with name
	GetWithName(name string) (MingleClient, error)

	// GetConnectedWithName returns MingleClient object with name and status is connected
	GetConnectedWithName(name string) (MingleClient, error)

	// GetAll returns all MingleClient
	GetAll() []MingleClient

	// GetAllConnected returns all MingleClient which status is connected
	GetAllConnected() []MingleClient

	// RegistryBeforAfterHandler registry BeforeStartHandle
	RegistryBeforAfterHandler(handler BeforeStartHandle)
}
