package types

import rtclient "sigs.k8s.io/controller-runtime/pkg/client"

// Predicate filters events before enqueuing the keys.
type Predicate interface {
	// Create returns true if the Create event should be processed
	Create(obj rtclient.Object) bool

	// Delete returns true if the Delete event should be processed
	Delete(obj rtclient.Object) bool

	// Update returns true if the Update event should be processed
	Update(oldObj, newObj rtclient.Object) bool

	// Generic returns true if the Generic event should be processed
	Generic(obj rtclient.Object) bool
}
