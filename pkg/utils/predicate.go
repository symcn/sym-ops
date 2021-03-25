package utils

import (
	"reflect"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// annotations need ignore key
var (
	ClusterAnnotationMonitor     = "k8s.io/monitor"
	ClusterAnnotationLoki        = "k8s.io/loki"
	WorkLoadAnnotationHpa        = "hpa.autoscaling.dmall.com/Hpa"
	WorkLoadAnnotationHpaMetrics = "hpa.autoscaling.dmall.com/Metrics"

	IgnoreAnnotationKey = []string{
		ClusterAnnotationMonitor,
		WorkLoadAnnotationHpa,
		WorkLoadAnnotationHpaMetrics,
	}
)

// ObjecteMetaEqual compare object is equal
// deepequal finalizer and labels
// filter annotations and then deepequal annotatins
func ObjecteMetaEqual(obj1 metav1.Object, obj2 metav1.Object) bool {
	if obj1 == nil && obj2 == nil {
		return true
	}
	if obj1 == nil || obj2 == nil {
		return false
	}

	// finalizer
	if !reflect.DeepEqual(obj1.GetFinalizers(), obj2.GetFinalizers()) {
		klog.V(4).Infof("Object %s/%s finalizer different.", obj1.GetNamespace(), obj1.GetName())
		return false
	}

	// labels
	if !reflect.DeepEqual(obj1.GetLabels(), obj2.GetLabels()) {
		klog.V(4).Infof("Object %s/%s labels different.", obj1.GetNamespace(), obj1.GetName())
		return false
	}

	// annotations
	// filter
	obj1Annotation := map[string]string{}
	for k, v := range obj1.GetAnnotations() {
		if !isIgnoreAnnotationKey(k) {
			obj1Annotation[k] = v
		}
	}
	obj2Annotation := map[string]string{}
	for k, v := range obj2.GetAnnotations() {
		if !isIgnoreAnnotationKey(k) {
			obj2Annotation[k] = v
		}
	}
	// compare
	if !reflect.DeepEqual(obj1Annotation, obj2Annotation) {
		klog.V(4).Infof("Object %s/%s annotations different.", obj1.GetNamespace(), obj1.GetName())
		return false
	}

	return true
}

func isIgnoreAnnotationKey(k string) bool {
	for _, key := range IgnoreAnnotationKey {
		if strings.EqualFold(key, k) {
			return true
		}
	}
	return false
}
