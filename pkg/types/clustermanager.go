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

// MingleClient mingle client
// wrap controller-runtime manager
type MingleClient interface {
	ResourceOperate

	// if dissatisfy can use this interface get Kubernetes resource
	KubernetesResource

	// if dissatisfy can use this interface get controller-runtime manager resource
	ControllerRuntimeManagerResource

	// Start client and blocks until the context is cancelled
	// Returns an error if there is an error starting
	Start(ctx context.Context) error

	// IsConnected return connected status
	IsConnected() bool
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

	//HasSynced return true if all informers underlying store has synced
	HasSynced() bool

	// Get retrieves an obj for the given object key from the Kubernetes Cluster with timeout.
	// obj must be a struct pointer so that obj can be updated with the response
	// returned by the Server.
	Get(key ktypes.NamespacedName, obj rtclient.Object)

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
