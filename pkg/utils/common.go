package utils

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// RemoveSliceString remove slice unexpect key
func RemoveSliceString(slice []string, s string) []string {
	if slice == nil {
		return nil
	}
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return result
}

// SliceContainsString return slice is contains str
func SliceContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// GetMapWithDefaultValue returns map specify key value, if not exist return default value
func GetMapWithDefaultValue(m map[string]string, key string, def string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

// MergeMap merge map which key and values both string
// if m2 have same key for m1, override m1
func MergeMap(m1 map[string]string, m2 map[string]string) map[string]string {
	merge := make(map[string]string)
	for k, v := range m1 {
		merge[k] = v
	}
	for k, v := range m2 {
		merge[k] = v
	}
	return merge
}

// TransformK8sName tranform kuberntes standerd name
// 1. replace all '_' to '-'
// 2. ToLower
func TransformK8sName(str string) string {
	return strings.ToLower(strings.ReplaceAll(str, "_", "-"))
}

// GetPodContainerImageVersion returns podSpec specify container image version
func GetPodContainerImageVersion(containerName string, podSpec *corev1.PodSpec) string {
	if podSpec == nil {
		return ""
	}

	for _, container := range podSpec.Containers {
		if strings.EqualFold(container.Name, containerName) {
			names := strings.Split(container.Image, ":")
			if len(names) == 2 {
				return names[1]
			}
		}
	}
	// if return this, means not found specify container
	return ""
}

// TransInt32Ptr2Int32 transform *int32 to int32 with default value.
func TransInt32Ptr2Int32(i *int32, def int32) int32 {
	if i == nil {
		return def
	}
	return *i
}
