package object

import (
	"bytes"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	// YAMLSeparator is a separator for multi-document YAML files.
	YAMLSeparator = "\n---\n"
)

// K8sObject is an in-memory representation of a k8s object, used for moving between different representations
// (Unstructured, JSON, YAML) with cached rendering.
type K8sObject struct {
	Group     string
	Kind      string
	Name      string
	Namespace string

	object *unstructured.Unstructured
	json   []byte
	yaml   []byte
}

// NewK8sObject creates a new K8sObject and returns a ptr to it.
func NewK8sObject(u *unstructured.Unstructured, json, yaml []byte) *K8sObject {
	o := &K8sObject{
		Name:      u.GetName(),
		Namespace: u.GetNamespace(),
		object:    u,
		json:      json,
		yaml:      yaml,
	}
	gvk := u.GetObjectKind().GroupVersionKind()
	o.Group = gvk.Group
	o.Kind = gvk.Kind
	return o
}

// ParseJSON2K8sObject parses JSON to an K8sObject.
func ParseJSON2K8sObject(json []byte) (*K8sObject, error) {
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

// ParseYAML2K8sObject parsed YAML to an K8sObject.
func ParseYAML2K8sObject(yaml []byte) (*K8sObject, error) {
	r := bytes.NewReader(yaml)
	decoder := k8syaml.NewYAMLOrJSONDecoder(r, 1024)
	out := &unstructured.Unstructured{}
	err := decoder.Decode(out)
	if err != nil {
		return nil, fmt.Errorf("error parsing yaml to unstructured object: %v", err)
	}
	return NewK8sObject(out, nil, yaml), nil
}

// UnstructuredObject returns the raw object, primarily for testing
func (k *K8sObject) UnstructuredObject() *unstructured.Unstructured {
	return k.object
}

// GroupVersionKind returns the GroupVersionKind for the K8sObject
func (k *K8sObject) GroupVersionKind() schema.GroupVersionKind {
	return k.object.GroupVersionKind()
}

// GroupKind returns the GroupKind for the K8sObject
func (k *K8sObject) GroupKind() schema.GroupKind {
	return k.object.GroupVersionKind().GroupKind()
}

// Hash returns a unique hash for the K8sObject
func (k *K8sObject) Hash() string {
	return hash(k.Kind, k.Namespace, k.Name)
}

// HashNameKind returns a hash for the K8sObject based on the name and kind only.
func (k *K8sObject) HashNameKind() string {
	return hashNameKind(k.Kind, k.Name)
}

// JSON returns a JSON representation of the K8sObject, using an internal cache.
func (k *K8sObject) JSON() ([]byte, error) {
	if k.json != nil {
		return k.json, nil
	}

	b, err := k.object.MarshalJSON()
	if err != nil {
		return nil, err
	}
	k.json = b
	return b, nil
}

func (k *K8sObject) YAML() ([]byte, error) {
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

// YAML2String returns a YAML representation of the K8sObject, or an error string if the K8sObject cannot be rendered to YAML.
func (k *K8sObject) YAML2String() string {
	y, err := k.YAML()
	if err != nil {
		return ""
	}
	return string(y)
}

func (k *K8sObject) AddLabels(labels map[string]string) {
	merged := make(map[string]string)
	for k, v := range k.object.GetLabels() {
		merged[k] = v
	}
	for k, v := range labels {
		merged[k] = v
	}
	k.object.SetLabels(merged)
	k.json = nil
	k.yaml = nil
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

// K8sObjects holds a conllection of k8s objects, so that we can filter / sequence them
type K8sObjects []*K8sObject

// K8sObjectsFromUnstructuredSlice returns an Objects ptr type from a slice of Unstructured.
func K8sObjectsFromUnstructuredSlice(objs []*unstructured.Unstructured) (K8sObjects, error) {
	var ret K8sObjects
	for _, o := range objs {
		ret = append(ret, NewK8sObject(o, nil, nil))
	}
	return ret, nil
}

// RenderTemplate render chart template to k8s object
func RenderTemplate(chartPkg []byte, rlsName, ns string, overrideValue string) (K8sObjects, error) {
	chrt, err := loader.LoadArchive(bytes.NewBuffer(chartPkg))
	if err != nil {
		return nil, fmt.Errorf("loading chart has an error: %v", err)
	}
	chrtVals, err := chartutil.ReadValues([]byte(overrideValue))
	if err != nil {
		return nil, fmt.Errorf("read overridevalue has an error: %v", err)
	}
	opts := chartutil.ReleaseOptions{
		Name:      rlsName,
		Namespace: ns,
		Revision:  1,
		IsInstall: true,
		IsUpgrade: false,
	}
	chrtValues, err := chartutil.ToRenderValues(chrt, chrtVals, opts, nil)
	if err != nil {
		return nil, fmt.Errorf("render chart values has an error: %v", err)
	}

	renderedTpls, err := engine.Render(chrt, chrtValues)
	if err != nil {
		return nil, fmt.Errorf("render error: %v", err)
	}

	var objects []*K8sObject
	for _, tpl := range renderedTpls {
		yaml := removeNonYAMLLines(tpl)
		if yaml == "" {
			continue
		}
		o, err := ParseYAML2K8sObject([]byte(yaml))
		if err != nil {
			klog.Errorf("Failed to parse yaml %s to k8s object: %v", yaml, err)
			continue
		}
		objects = append(objects, o)
		klog.V(5).Infof("Render k8s object %s %s/%s success", o.Kind, o.Namespace, o.Name)
	}
	return objects, nil
}

func removeNonYAMLLines(yamlStr string) string {
	out := ""
	for _, s := range strings.Split(yamlStr, "\n") {
		if strings.HasPrefix(s, "#") {
			continue
		}
		out += s + "\n"
	}

	return strings.TrimSpace(out)
}
