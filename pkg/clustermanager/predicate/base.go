package predicate

import (
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type base struct {
	handler func(obj rtclient.Object) bool
}

// Create returns true if the Create event should be processed
func (b *base) Create(obj rtclient.Object) bool {
	return b.handler(obj)
}

// Delete returns true if the Delete event should be processed
func (b *base) Delete(obj rtclient.Object) bool {
	return b.handler(obj)
}

// Update returns true if the Update event should be processed
func (b *base) Update(objObj, newObj rtclient.Object) bool {
	return b.handler(newObj)
}

// Generic returns true if the Generic event should be processed
func (b *base) Generic(obj rtclient.Object) bool {
	return b.handler(obj)
}
