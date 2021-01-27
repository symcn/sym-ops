package types

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

// Reconciler interface, define Reconcile handle
type Reconciler interface {
	Reconcile(req types.NamespacedName) (requeue NeedRequeue, after time.Duration, err error)
}
