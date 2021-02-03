package handler

import (
	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceEventHandler struct {
	Queue        types.WorkQueue
	EventHandler types.EventHandler
	Predicates   []types.Predicate
}

// NewResourceEventHandler build controller-runtime ResourceEventHandler
func NewResourceEventHandler(queue types.WorkQueue, handler types.EventHandler, predicates ...types.Predicate) cache.ResourceEventHandler {
	return &resourceEventHandler{Queue: queue, EventHandler: handler, Predicates: predicates}
}

// OnAdd is called when an object is added.
func (e *resourceEventHandler) OnAdd(obj interface{}) {
	o, ok := obj.(rtclient.Object)
	if !ok {
		klog.Errorf("OnAdd missing Object[%T] %v", obj, obj)
		return
	}

	for _, p := range e.Predicates {
		if !p.Create(o) {
			return
		}
	}

	e.EventHandler.Create(o, e.Queue)
}

// OnUpdate is called when an object is modified. Note that oldObj is the
// last known state of the object-- it is possible that several changes
// were combined together, so you can't use this to see every single
// change. OnUpdate is also called when a re-list happens, and it will
// get called even if nothing changed. This is useful for periodically
// evaluating or syncing something.
func (e *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	o, ok := oldObj.(rtclient.Object)
	if !ok {
		klog.Errorf("OnUpdate missing ObjectOld[%T] %v", oldObj, oldObj)
		return
	}

	n, ok := newObj.(rtclient.Object)
	if !ok {
		klog.Errorf("OnUpdate missing ObjectNew[%T] %v", newObj, newObj)
		return
	}
	for _, p := range e.Predicates {
		if !p.Update(o, n) {
			return
		}
	}
	e.EventHandler.Update(o, n, e.Queue)
}

// OnDelete will get the final state of the item if it is known, otherwise
// it will get an object of type DeletedFinalStateUnknown. This can
// happen if the watch is closed and misses the delete event and we don't
// notice the deletion until the subsequent re-list.
func (e *resourceEventHandler) OnDelete(obj interface{}) {
	// Deal with tombstone events by pulling the object out.  Tombstone events wrap the object in a
	// DeleteFinalStateUnknown struct, so the object needs to be pulled out.
	// Copied from sample-controller
	// This should never happen if we aren't missing events, which we have concluded that we are not
	// and made decisions off of this belief.  Maybe this shouldn't be here?
	if _, ok := obj.(rtclient.Object); !ok {
		// If the object doesn't have Metadata, assume it is a tombstone object of type DeletedFinalStateUnknow
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			klog.Errorf("Error decoding objectes. Expected cache.DeletedFinalStateUnknow Object[%T] %v", obj, obj)
			return
		}
		// Set obj to the tombstone obj
		obj = tombstone.Obj
	}

	// Pull Object out of the object
	o, ok := obj.(rtclient.Object)
	if !ok {
		klog.Errorf("OnDelete missing Object[%T] %v", obj, obj)
		return
	}

	for _, p := range e.Predicates {
		if !p.Delete(o) {
			return
		}
	}

	// Invoke delete handler
	e.EventHandler.Delete(o, e.Queue)
}
