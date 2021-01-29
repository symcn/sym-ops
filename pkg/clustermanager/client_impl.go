package clustermanager

import (
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	rtcache "sigs.k8s.io/controller-runtime/pkg/cache"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	rtmanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

// GetInformer fetches or constructs an informer for the given object that corresponds to a single
// API kind and resource.
func (c *client) GetInformer(obj rtclient.Object) (rtcache.Informer, error) {
	panic("not implemented") // TODO: Implement
}

// AddResourceEventHandler
// 1. GetInformer
// 2. Adds an event handler to the shared informer using the shared informer's resync
//	period.  Events to a single handler are delivered sequentially, but there is no coordination
//	between different handlers.
func (c *client) AddResourceEventHandler(obj rtclient.Object, handler cache.ResourceEventHandler) error {
	panic("not implemented") // TODO: Implement
}

//HasSynced return true if all informers underlying store has synced
func (c *client) HasSynced() bool {
	panic("not implemented") // TODO: Implement
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster with timeout.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (c *client) Get(key ktypes.NamespacedName, obj rtclient.Object) {
	panic("not implemented") // TODO: Implement
}

// Create saves the object obj in the Kubernetes cluster with timeout.
func (c *client) Create(obj rtclient.Object, opts ...rtclient.CreateOption) error {
	panic("not implemented") // TODO: Implement
}

// Delete deletes the given obj from Kubernetes cluster with timeout.
func (c *client) Delete(obj rtclient.Object, opts ...rtclient.DeleteOption) error {
	panic("not implemented") // TODO: Implement
}

// Update updates the given obj in the Kubernetes cluster with timeout. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *client) Update(obj rtclient.Object, opts ...rtclient.UpdateOption) error {
	panic("not implemented") // TODO: Implement
}

// Update updates the fields corresponding to the status subresource for the
// given obj with timeout. obj must be a struct pointer so that obj can be updated
// with the content returned by the Server.
func (c *client) StatusUpdate(obj rtclient.Object, opts ...rtclient.UpdateOption) error {
	panic("not implemented") // TODO: Implement
}

// Patch patches the given obj in the Kubernetes cluster with timeout. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *client) Patch(obj rtclient.Object, patch rtclient.Patch, opts ...rtclient.PatchOption) error {
	panic("not implemented") // TODO: Implement
}

// DeleteAllOf deletes all objects of the given type matching the given options with timeout.
func (c *client) DeleteAllOf(obj rtclient.Object, opts ...rtclient.DeleteAllOfOption) error {
	panic("not implemented") // TODO: Implement
}

// GetRestConfig return Kubernetes rest Config
func (c *client) GetKubeRestConfig() *rest.Config {
	panic("not implemented") // TODO: Implement
}

// GetKubeInterface return Kubernetes Interface.
// kubernetes.ClientSet impl kubernetes.Interface
func (c *client) GetKubeInterface() kubernetes.Interface {
	panic("not implemented") // TODO: Implement
}

// GetCtrlRtManager return controller-runtime manager object
func (c *client) GetCtrlRtManager() rtmanager.Manager {
	panic("not implemented") // TODO: Implement
}

// GetCtrlRtCache return controller-runtime cache object
func (c *client) GetCtrlRtCache() rtcache.Cache {
	panic("not implemented") // TODO: Implement
}

// GetCtrlRtClient return controller-runtime client
func (c *client) GetCtrlRtClient() rtclient.Client {
	panic("not implemented") // TODO: Implement
}
