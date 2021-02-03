package predicate

import (
	"strings"

	"github.com/symcn/sym-ops/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type predicateNamespace struct {
	watchNamespaceList []string
}

// NamespacePredicate filter namespace
func NamespacePredicate(nslist ...string) types.Predicate {
	return &base{
		handler: func(obj client.Object) bool {
			for _, ns := range nslist {
				if ns == "*" || strings.EqualFold(ns, obj.GetNamespace()) {
					return true
				}
			}
			return false
		},
	}
}

// LabelsKeyPredicate filter labels key not exists
func LabelsKeyPredicate(keys ...string) types.Predicate {
	return &base{
		handler: func(obj client.Object) bool {
			if len(obj.GetLabels()) == 0 {
				return false
			}
			for _, key := range keys {
				if _, ok := obj.GetLabels()[key]; !ok {
					return false
				}
			}
			return true
		},
	}
}
