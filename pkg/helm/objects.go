package helm

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/symcn/sym-ops/pkg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

// K8sObject is an in-memory representation of a k8s object, used for moving between different representations
// (Unstructured, JSON, YAML) with cached rendering.
type K8sObject interface {
	// GetName returns k8sobj name
	GetName() string

	// GetNamespace returns k8sobj namespace
	GetNamespace() string

	// GetLabels returns k8sobj labels
	GetLabels() map[string]string

	// UnstructuredObject returns the raw object, primarily for testing
	UnstructuredObject() *unstructured.Unstructured

	// GroupVersionKind returns the GroupVersionKind for the k8sobj
	GroupVersionKind() schema.GroupVersionKind

	// GroupKind returns the GroupKind for the k8sobj
	GroupKind() schema.GroupKind

	// Hash returns a unique hash for the k8sobj
	Hash() string

	// HashNameKind returns a hash for the k8sobj based on the name and kind only.
	HashNameKind() string

	// JSON returns a JSON representation of the k8sobj, using an internal cache.
	JSON() ([]byte, error)

	// YAML returns a yaml representation of the k8sobj, using an internal cache.
	YAML() ([]byte, error)

	// YAML2String returns a YAML representation of the k8sobj, or an error string if the k8sobj cannot be rendered to YAML.
	YAML2String() string

	// AddLabels add labels to k8sobj
	AddLabels(labels map[string]string)
}

type k8sobj struct {
	group     string
	kind      string
	name      string
	namespace string

	object *unstructured.Unstructured
	json   []byte
	yaml   []byte
}

// NewK8sObject creates a new k8sobj and returns a ptr to it.
func NewK8sObject(u *unstructured.Unstructured, json, yaml []byte) K8sObject {
	o := &k8sobj{
		name:      u.GetName(),
		namespace: u.GetNamespace(),
		object:    u,
		json:      json,
		yaml:      yaml,
	}
	gvk := u.GetObjectKind().GroupVersionKind()
	o.group = gvk.Group
	o.kind = gvk.Kind
	return o
}

// GetName returns k8sobj name
func (k *k8sobj) GetName() string {
	return k.name
}

// GetNamespace returns k8sobj namespace
func (k *k8sobj) GetNamespace() string {
	return k.namespace
}

// GetLabels returns k8sobj labels
func (k *k8sobj) GetLabels() map[string]string {
	return k.object.GetLabels()
}

// UnstructuredObject returns the raw object, primarily for testing
func (k *k8sobj) UnstructuredObject() *unstructured.Unstructured {
	return k.object
}

// GroupVersionKind returns the GroupVersionKind for the k8sobj
func (k *k8sobj) GroupVersionKind() schema.GroupVersionKind {
	return k.object.GroupVersionKind()
}

// GroupKind returns the GroupKind for the k8sobj
func (k *k8sobj) GroupKind() schema.GroupKind {
	return k.object.GroupVersionKind().GroupKind()
}

// Hash returns a unique hash for the k8sobj
func (k *k8sobj) Hash() string {
	return hash(k.kind, k.namespace, k.name)
}

// HashNameKind returns a hash for the k8sobj based on the name and kind only.
func (k *k8sobj) HashNameKind() string {
	return hashNameKind(k.kind, k.name)
}

// JSON returns a JSON representation of the k8sobj, using an internal cache.
func (k *k8sobj) JSON() ([]byte, error) {
	if k.json != nil {
		return k.json, nil
	}

	if k.object == nil {
		return nil, errors.New("object is nil")
	}
	b, err := k.object.MarshalJSON()
	if err != nil {
		return nil, err
	}
	k.json = b
	return b, nil
}

// YAML returns a yaml representation of the k8sobj, using an internal cache.
func (k *k8sobj) YAML() ([]byte, error) {
	if k.yaml != nil {
		return k.yaml, nil
	}

	j, err := k.JSON()
	if err != nil {
		return nil, err
	}

	y, err := yaml.JSONToYAML(j)
	if err != nil {
		return nil, err
	}
	k.yaml = y
	return y, nil
}

// YAML2String returns a YAML representation of the k8sobj, or an error string if the k8sobj cannot be rendered to YAML.
func (k *k8sobj) YAML2String() string {
	y, err := k.YAML()
	if err != nil {
		return ""
	}
	return string(y)
}

// AddLabels add labels to k8sobj
func (k *k8sobj) AddLabels(labels map[string]string) {
	k.object.SetLabels(utils.MergeMap(k.object.GetLabels(), labels))
	k.json = nil
	k.yaml = nil
}

// ParseYAML2K8sObject parsed YAML to an k8sobj.
func ParseYAML2K8sObject(yaml []byte) (K8sObject, error) {
	r := bytes.NewReader(yaml)
	decoder := k8syaml.NewYAMLOrJSONDecoder(r, 1024)
	out := &unstructured.Unstructured{}
	err := decoder.Decode(out)
	if err != nil {
		return nil, fmt.Errorf("error parsing yaml to unstructured object: %v", err)
	}
	return NewK8sObject(out, nil, yaml), nil
}

// ParseJSON2K8sObject parses JSON to an k8sobj.
func ParseJSON2K8sObject(json []byte) (K8sObject, error) {
	o, _, err := unstructured.UnstructuredJSONScheme.Decode(json, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error parsing json to unstructured object: %v", err)
	}

	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("parsed unexpected type %T", o)
	}
	return NewK8sObject(u, json, nil), nil
}

// hash returns a unique, insecure hash based on kind, namespace and name.
func hash(kind, namespace, name string) string {
	switch kind {
	case "ClusterRole", "ClusterRoleBinding":
		namespace = ""
	}
	return strings.Join([]string{kind, namespace, name}, ":")
}

// hashNameKind returns a unique, insecure hash based on kind and name.
func hashNameKind(kind, name string) string {
	return strings.Join([]string{kind, name}, ":")
}
