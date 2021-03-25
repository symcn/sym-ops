package helm

import (
	"reflect"
	"strings"
	"testing"

	"github.com/symcn/sym-ops/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	replicas int32 = 2
	label1         = map[string]string{
		"l1": "v1",
		"l2": "v2",
	}
	label2 = map[string]string{
		"l1": "v2",
		"l2": "v1",
		"l3": "v3",
	}

	deploy = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "my-dnstools",
			Labels: label1,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"run": "my-dnstools",
				},
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run": "my-dnstools",
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "dnstools",
							Image: "infoblox/dnstools:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
								{
									ContainerPort: 443,
								},
							},
							Command: []string{"sleep", "36000"},
						},
					},
				},
			},
		},
	}

	yamlStr = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-dnstools
spec:
  selector:
    matchLabels:
      run: my-dnstools
  replicas: 1
  template:
    metadata:
      labels:
        run: my-dnstools
      annotations:
        sidecar.istio.io/inject: 'true'
    spec:
      containers:
      - name: dnstools
        image: infoblox/dnstools:latest
        ports:
        - containerPort: 80
        - containerPort: 443
        command: ["sleep", "36000"]
`
	jsonStr = `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"my-dnstools"},"spec":{"replicas":1,"selector":{"matchLabels":{"run":"my-dnstools"}},"template":{"metadata":{"annotations":{"sidecar.istio.io/inject":"true"},"labels":{"run":"my-dnstools"}},"spec":{"containers":[{"command":["sleep","36000"],"image":"infoblox/dnstools:latest","name":"dnstools","ports":[{"containerPort":80},{"containerPort":443}]}]}}}}`
)

func TestNewK8sObject(t *testing.T) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deploy)
	if err != nil {
		t.Error(err)
		return
	}

	k8sObj := NewK8sObject(&unstructured.Unstructured{Object: obj}, nil, nil)
	t.Log(k8sObj.UnstructuredObject())
	t.Log(k8sObj.GroupVersionKind().String())
	t.Log(k8sObj.GroupKind().String())
	t.Log(k8sObj.Hash())
	t.Log(k8sObj.HashNameKind())

	k8sObj.AddLabels(label2)
	if !reflect.DeepEqual(k8sObj.GetLabels(), utils.MergeMap(label1, label2)) {
		t.Errorf("AddLabels test failed: raw:%v, add:%v, expect:%v but got %v", label1, label2, utils.MergeMap(label1, label2), k8sObj.GetLabels())
		return
	}

	// error
	_, err = ParseYAML2K8sObject([]byte(k8sObj.YAML2String()))
	if err == nil {
		t.Error("not kind must be error")
		return
	}

	d, err := k8sObj.JSON()
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseJSON2K8sObject(d)
	if err == nil {
		t.Error("not kind must be error")
		return
	}
}

func TestParseYAML2K8sObject(t *testing.T) {
	_, err := ParseYAML2K8sObject([]byte("error yaml"))
	if err == nil {
		t.Errorf("ParseJSON2K8sObject error yaml string must be error")
		return
	}

	_, err = ParseYAML2K8sObject([]byte(yamlStr))
	if err != nil {
		t.Error(err)
		return
	}

	obj := &k8sobj{}
	if obj.YAML2String() != "" {
		t.Errorf("YAML2String panic test failed: want '' but got %s", obj.YAML2String())
		return
	}
}

func TestParseJSON2K8sObject(t *testing.T) {
	_, err := ParseJSON2K8sObject([]byte("error json"))
	if err == nil {
		t.Errorf("ParseJSON2K8sObject error yaml string must be error")
		return
	}

	_, err = ParseJSON2K8sObject([]byte(jsonStr))
	if err != nil {
		t.Error(err)
		return
	}
}

func TestHash(t *testing.T) {
	kind := "ClusterRole"
	namespace := "ns"
	name := "name"
	if hash(kind, namespace, name) != strings.Join([]string{kind, "", name}, ":") {
		t.Error("hash test error")
		return
	}
}
