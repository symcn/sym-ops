package clustermanager

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

// GetInformer fetches or constructs an informer for the given object that corresponds to a single
// API kind and resource.
func (c *client) GetInformer(obj rtclient.Object) (rtcache.Informer, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	informer, err := c.ctrlRtCache.GetInformer(ctx, obj)
	if err != nil {
		return nil, err
	}
	c.informerList = append(c.informerList, informer)
	return informer, nil
}

// AddResourceEventHandler
// 1. GetInformer
// 2. Adds an event handler to the shared informer using the shared informer's resync
//	period.  Events to a single handler are delivered sequentially, but there is no coordination
//	between different handlers.
func (c *client) AddResourceEventHandler(obj rtclient.Object, handler cache.ResourceEventHandler) error {
	informer, err := c.GetInformer(obj)
	if err != nil {
		return err
	}
	informer.AddEventHandler(handler)
	return nil
}

// IndexFields adds an index with the given field name on the given object type
// by using the given function to extract the value for that field.  If you want
// compatibility with the Kubernetes API server, only return one key, and only use
// fields that the API server supports.  Otherwise, you can return multiple keys,
// and "equality" in the field selector means that at least one key matches the value.
// The FieldIndexer will automatically take care of indexing over namespace
// and supporting efficient all-namespace queries.
func (c *client) SetIndexField(obj rtclient.Object, field string, extractValue rtclient.IndexerFunc) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtManager.GetFieldIndexer().IndexField(ctx, obj, field, extractValue)
}

// HasSynced return true if all informers underlying store has synced
// !import if informerlist is empty, will return true
func (c *client) HasSynced() bool {
	if !c.started {
		// if not start, the informer will not synced
		return false
	}

	for _, informer := range c.informerList {
		if !informer.HasSynced() {
			return false
		}
	}
	return true
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster with timeout.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (c *client) Get(key ktypes.NamespacedName, obj rtclient.Object) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.Get(ctx, key, obj)
}

// Create saves the object obj in the Kubernetes cluster with timeout.
func (c *client) Create(obj rtclient.Object, opts ...rtclient.CreateOption) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.Create(ctx, obj, opts...)
}

// Delete deletes the given obj from Kubernetes cluster with timeout.
func (c *client) Delete(obj rtclient.Object, opts ...rtclient.DeleteOption) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.Delete(ctx, obj, opts...)
}

// Update updates the given obj in the Kubernetes cluster with timeout. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *client) Update(obj rtclient.Object, opts ...rtclient.UpdateOption) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.Update(ctx, obj, opts...)
}

// Update updates the fields corresponding to the status subresource for the
// given obj with timeout. obj must be a struct pointer so that obj can be updated
// with the content returned by the Server.
func (c *client) StatusUpdate(obj rtclient.Object, opts ...rtclient.UpdateOption) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.Status().Update(ctx, obj, opts...)
}

// Patch patches the given obj in the Kubernetes cluster with timeout. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (c *client) Patch(obj rtclient.Object, patch rtclient.Patch, opts ...rtclient.PatchOption) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.Patch(ctx, obj, patch, opts...)
}

// DeleteAllOf deletes all objects of the given type matching the given options with timeout.
func (c *client) DeleteAllOf(obj rtclient.Object, opts ...rtclient.DeleteAllOfOption) error {
	ctx, cancel := context.WithTimeout(context.TODO(), c.ExecTimeout)
	defer cancel()

	return c.ctrlRtClient.DeleteAllOf(ctx, obj, opts...)
}

// GetRestConfig return Kubernetes rest Config
func (c *client) GetKubeRestConfig() *rest.Config {
	return c.kubeRestConfig
}

// GetKubeInterface return Kubernetes Interface.
// kubernetes.ClientSet impl kubernetes.Interface
func (c *client) GetKubeInterface() kubernetes.Interface {
	return c.kubeInterface
}

// GetCtrlRtManager return controller-runtime manager object
func (c *client) GetCtrlRtManager() rtmanager.Manager {
	return c.ctrlRtManager
}

// GetCtrlRtCache return controller-runtime cache object
func (c *client) GetCtrlRtCache() rtcache.Cache {
	return c.ctrlRtCache
}

// GetCtrlRtClient return controller-runtime client
func (c *client) GetCtrlRtClient() rtclient.Client {
	return c.ctrlRtClient
}
