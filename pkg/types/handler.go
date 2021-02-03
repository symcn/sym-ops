package types

import rtclient "sigs.k8s.io/controller-runtime/pkg/client"

// ObjectTransformFunc EventHandler transform object
type ObjectTransformFunc func(obj rtclient.Object) string

// EventHandler deal event handler
type EventHandler interface {
	// Create is called in response to an create event - e.g. Pod Creation.
	Create(obj rtclient.Object, queue WorkQueue)

	// Update is called in response to an update event -  e.g. Pod Updated.
	Update(oldObj, newObj rtclient.Object, queue WorkQueue)

	// Delete is called in response to a delete event - e.g. Pod Deleted.
	Delete(obj rtclient.Object, queue WorkQueue)

	// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
	// external trigger request - e.g. reconcile Autoscaling, or a Webhook.
	Generic(obj rtclient.Object, queue WorkQueue)
}
