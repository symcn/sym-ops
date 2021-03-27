package patch

import (
	"fmt"
	"reflect"
	"unsafe"

	json "github.com/json-iterator/go"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

func init() {
	// k8s.io/apimachinery/pkg/util/intstr.IntOrString behaves really badly
	// from JSON marshaling point of view, it can't be empty basically.
	// So we need to override the defined marshaling behaviour and write nil
	// instead of 0, because usually (in all observed cases) 0 means "not set"
	// for IntOrStr types.
	// To make this happen we need to pull in json-iterator and override the
	// factory marshaling overrides.
	json.RegisterTypeEncoderFunc("intstr.IntOrString",
		func(ptr unsafe.Pointer, stream *json.Stream) {
			i := (*intstr.IntOrString)(ptr)
			if i.IntValue() == 0 {
				if i.StrVal != "" && i.StrVal != "0" {
					stream.WriteString(i.StrVal)
				} else {
					stream.WriteNil()
				}
			}
		},
		func(ptr unsafe.Pointer) bool {
			i := (*intstr.IntOrString)(ptr)
			return i.IntValue() == 0 && (i.StrVal == "" || i.StrVal == "0")
		},
	)
}

func DeleteNullInJson(jsonBytes []byte) ([]byte, map[string]interface{}, error) {
	var patchMap map[string]interface{}

	err := json.ConfigCompatibleWithStandardLibrary.Unmarshal(jsonBytes, &patchMap)
	if err != nil {
		return nil, nil, fmt.Errorf("counld not unmarshal json patch: %v", err)
	}

	filteredMap := deleteNullInObj(patchMap)
	o, err := json.ConfigCompatibleWithStandardLibrary.Marshal(filteredMap)
	if err != nil {
		return nil, nil, fmt.Errorf("could not marshal filtered patch map: %v", err)
	}
	return o, filteredMap, nil
}

func deleteNullInObj(m map[string]interface{}) map[string]interface{} {
	filterdMap := make(map[string]interface{})

	for k, v := range m {
		if v == nil || isZero(reflect.ValueOf(v)) {
			continue
		}

		switch typVal := v.(type) {
		case []interface{}:
			if len(typVal) > 0 {
				filterdMap[k] = deleteNullInSlice(typVal)
			}
		case string, float64, bool, int64, nil:
			filterdMap[k] = typVal
		case map[string]interface{}:
			if len(typVal) == 0 {
				filterdMap[k] = typVal
				continue
			}

			var filteredSubMap map[string]interface{}
			filteredSubMap = deleteNullInObj(typVal)
			if len(filteredSubMap) != 0 {
				filterdMap[k] = filteredSubMap
			}
		default:
			filterdMap[k] = typVal
		}
	}
	return filterdMap
}

func deleteNullInSlice(m []interface{}) []interface{} {
	filteredSlice := make([]interface{}, 0, len(m))

	for _, v := range m {
		if v == nil || isZero(reflect.ValueOf(v)) {
			continue
		}

		switch typVal := v.(type) {
		case []interface{}:
			if len(typVal) > 0 {
				filteredSlice = append(filteredSlice, deleteNullInSlice(typVal)...)
			}
		case string, float64, bool, int64, nil:
			filteredSlice = append(filteredSlice, v)
		case map[string]interface{}:
			filteredSlice = append(filteredSlice, deleteNullInObj(typVal))
		default:
			filteredSlice = append(filteredSlice, v)
		}
	}
	return filteredSlice
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	defer func() {
		if err := recover(); err != nil {
			// if have panic return not zero
			klog.Errorf("judge isZero panic:%v", err)
		}
	}()

	switch v.Kind() {
	case reflect.Float64, reflect.Int64, reflect.Bool:
		// just use it with json result, only have float64、int64 and bool
		return false
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	default:
		z := reflect.Zero(v.Type())
		return v.Interface() == z.Interface()
	}
}

func IgnoreStatusFields() CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteStatusField(current)
		if err != nil {
			return nil, nil, fmt.Errorf("delete current status field failed: %v", err)
		}
		modified, err = deleteStatusField(modified)
		if err != nil {
			return nil, nil, fmt.Errorf("delete modify status field failed: %v", err)
		}
		return current, modified, nil
	}
}

func deleteStatusField(obj []byte) ([]byte, error) {
	var objectMap map[string]interface{}
	err := json.ConfigCompatibleWithStandardLibrary.Unmarshal(obj, &objectMap)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal byte to map[string]interface{}: %v", err)
	}
	delete(objectMap, "status")
	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(objectMap)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal map[string]interface{} to byte: %v", err)
	}
	return obj, nil
}

func IgnoreVolumeClaimTemplateTypeMetaAndStatus() CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteVolumeClaimTemplateFields(current)
		if err != nil {
			return nil, nil, fmt.Errorf("delete current volumeclaimtemplate field failed: %v", err)
		}
		modified, err = deleteVolumeClaimTemplateFields(modified)
		if err != nil {
			return nil, nil, fmt.Errorf("delete modified volumeclaimtemplate field failed: %v", err)
		}
		return current, modified, nil
	}
}

func deleteVolumeClaimTemplateFields(obj []byte) ([]byte, error) {
	sts := appsv1.StatefulSet{}
	err := json.ConfigCompatibleWithStandardLibrary.Unmarshal(obj, &sts)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal byte to statefulset: %v", err)
	}
	for i := range sts.Spec.VolumeClaimTemplates {
		sts.Spec.VolumeClaimTemplates[i].Kind = ""
		sts.Spec.VolumeClaimTemplates[i].APIVersion = ""
		if sts.Spec.VolumeClaimTemplates[i].Spec.VolumeMode == nil {
			fs := corev1.PersistentVolumeFilesystem
			sts.Spec.VolumeClaimTemplates[i].Spec.VolumeMode = &fs
		}
		sts.Spec.VolumeClaimTemplates[i].Status = corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimPending,
		}
	}
	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(sts)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal statefulset to byte: %v", err)
	}
	return obj, nil
}

func IgnoreDeployReplicasFields() CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteDeployReplicasFields(current)
		if err != nil {
			return nil, nil, fmt.Errorf("delete current deploy replicas field failed: %v", err)
		}
		modified, err = deleteDeployReplicasFields(modified)
		if err != nil {
			return nil, nil, fmt.Errorf("delete modified deploy replicas field failed: %v", err)
		}
		return current, modified, nil
	}
}

func deleteDeployReplicasFields(obj []byte) ([]byte, error) {
	deploy := appsv1.Deployment{}
	err := json.ConfigCompatibleWithStandardLibrary.Unmarshal(obj, &deploy)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal byte to deployment: %v", err)
	}

	deploy.Spec.Replicas = nil

	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(deploy)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal deployment to byte: %v", err)
	}
	return obj, nil
}

func IgnoreStatefulSetReplicasFields() CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := deleteStatefulSetReplicasFields(current)
		if err != nil {
			return nil, nil, fmt.Errorf("delete current statefulset replicas field failed: %v", err)
		}
		modified, err = deleteStatefulSetReplicasFields(modified)
		if err != nil {
			return nil, nil, fmt.Errorf("delete modified statefulset replicas field failed: %v", err)
		}
		return current, modified, nil
	}
}

func deleteStatefulSetReplicasFields(obj []byte) ([]byte, error) {
	sts := appsv1.StatefulSet{}
	err := json.ConfigCompatibleWithStandardLibrary.Unmarshal(obj, &sts)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal byte to statefulset: %v", err)
	}

	sts.Spec.Replicas = nil

	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(sts)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal statefulset to byte: %v", err)
	}
	return obj, nil
}
