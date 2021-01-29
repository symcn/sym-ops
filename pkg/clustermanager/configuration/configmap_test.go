package configuration

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewClusterCfgManagerWithCM(t *testing.T) {
	kubeInterface := fake.NewSimpleClientset()

	cfg := NewClusterCfgManagerWithCM(kubeInterface, "", map[string]string{"k1": "v1"}, "data", "status")
	_, err := cfg.GetAll()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestConfigmap2ClusterCfgInfo(t *testing.T) {
	dataKey := "kubeconfig"
	statusKey := "status"

	cmlist := &v1.ConfigMapList{
		Items: []v1.ConfigMap{
			{
				Data: map[string]string{
					dataKey:   "data1",
					statusKey: "",
				},
			},
			{
				Data: map[string]string{
					dataKey: "data2",
				},
			},
			{
				Data: map[string]string{
					dataKey:   "data3",
					statusKey: "true",
				},
			},
			{
				Data: map[string]string{
					dataKey:   "data4",
					statusKey: "false",
				},
			},
			{
				Data: map[string]string{
					"":        "data5",
					statusKey: "true",
				},
			},
			{
				Data: map[string]string{
					"": "data6",
				},
			},
		},
	}

	list := Configmap2ClusterCfgInfo(cmlist, dataKey, statusKey)
	if len(list) != 2 {
		t.Errorf("expect return 2 list, but got %d", len(list))
		return
	}
	if list[0].GetKubeConfig() != "data2" {
		t.Errorf("expect data2 but got %+v", list[0])
	}
	if list[1].GetKubeConfig() != "data3" {
		t.Errorf("expect data3 but got %+v", list[1])
	}
}
