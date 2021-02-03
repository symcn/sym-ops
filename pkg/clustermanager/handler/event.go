package handler

import (
	"github.com/symcn/sym-ops/pkg/types"
	ktypes "k8s.io/apimachinery/pkg/types"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type transformNameEventHandler struct {
	NameFunc      types.ObjectTransformFunc
	NamespaceFunc types.ObjectTransformFunc
}

func (t *transformNameEventHandler) Create(obj rtclient.Object, queue types.WorkQueue) {
	queue.Add(ktypes.NamespacedName{
		Name:      t.NameFunc(obj),
		Namespace: t.NamespaceFunc(obj),
	})
}

func (t *transformNameEventHandler) Update(oldObj, newObj rtclient.Object, queue types.WorkQueue) {
	queue.Add(ktypes.NamespacedName{
		Name:      t.NameFunc(newObj),
		Namespace: t.NamespaceFunc(newObj),
	})
}

func (t *transformNameEventHandler) Delete(obj rtclient.Object, queue types.WorkQueue) {
	queue.Add(ktypes.NamespacedName{
		Name:      t.NameFunc(obj),
		Namespace: t.NamespaceFunc(obj),
	})
}

func (t *transformNameEventHandler) Generic(obj rtclient.Object, queue types.WorkQueue) {
	queue.Add(ktypes.NamespacedName{
		Name:      t.NameFunc(obj),
		Namespace: t.NamespaceFunc(obj),
	})
}
