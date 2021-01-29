package clusterconfiguration

import (
	"fmt"
	"testing"

	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	mockKubeConfigDir  = "/tmp/mockkube"
	mockKubeConfigPath = "/tmp/mockkube/kubeconfig%d.%s"
	suffix             = "yaml"
)

func TestNewClusterCfgManagerWithPath(t *testing.T) {
	t.Run("path not exist", func(t *testing.T) {
		_, err := NewClusterCfgManagerWithPath("aaa", suffix, types.KubeConfigTypeFile)
		if err == nil {
			t.Error("path not exist should be error")
			return
		}
	})

	t.Run("file", func(t *testing.T) {
		_, err := NewClusterCfgManagerWithPath("/etc/hosts", suffix, types.KubeConfigTypeFile)
		if err == nil {
			t.Error("file not support should be error")
			return
		}
	})

	t.Run("open dir fail", func(t *testing.T) {
		obj, err := NewClusterCfgManagerWithPath("/etc", suffix, types.KubeConfigTypeFile)
		if err != nil {
			t.Error("file not support should be error")
			return
		}

		cfg := obj.(*cfgWithPath)
		cfg.dir = "aaaaa"
		_, err = cfg.GetAll()
		if err == nil {
			t.Error("dir not exist should be error")
			return
		}
	})

	t.Run("not found suffix file", func(t *testing.T) {
		num := 3
		buildTmpWithMockKubeConfig(num)

		cfg, err := NewClusterCfgManagerWithPath(mockKubeConfigDir, "aaaaa", types.KubeConfigTypeFile)
		if err != nil {
			t.Error(err)
			return
		}
		list, err := cfg.GetAll()
		if err != nil {
			t.Error(err)
			return
		}
		if len(list) != 0 {
			t.Error("not exist suffix file, should get empty")
		}
	})

	t.Run("unsupport type file", func(t *testing.T) {
		num := 3
		buildTmpWithMockKubeConfig(num)

		cfg, err := NewClusterCfgManagerWithPath(mockKubeConfigDir, suffix, "unsupport kubeconfigtype")
		if err != nil {
			t.Error(err)
			return
		}
		list, err := cfg.GetAll()
		if err != nil {
			t.Error(err)
			return
		}
		if len(list) != 0 {
			t.Error("unsupport type file, should get empty")
		}
	})

	t.Run("normal type file", func(t *testing.T) {
		num := 3
		buildTmpWithMockKubeConfig(num)

		cfg, err := NewClusterCfgManagerWithPath(mockKubeConfigDir, suffix, types.KubeConfigTypeFile)
		if err != nil {
			t.Error(err)
			return
		}
		list, err := cfg.GetAll()
		if err != nil {
			t.Error(err)
			return
		}
		if len(list) != num {
			t.Errorf("build %d config file but got %d", num, len(list))
		}
	})

	t.Run("normal type rawstring", func(t *testing.T) {
		num := 3
		buildTmpWithMockKubeConfig(num)

		cfg, err := NewClusterCfgManagerWithPath(mockKubeConfigDir, suffix, types.KubeConfigTypeRawString)
		if err != nil {
			t.Error(err)
			return
		}
		list, err := cfg.GetAll()
		if err != nil {
			t.Error(err)
			return
		}
		if len(list) != num {
			t.Errorf("build %d config file but got %d", num, len(list))
		}
	})
}

func buildTmpWithMockKubeConfig(num int) {
	for i := 0; i < num; i++ {
		cfg := clientcmdapi.NewConfig()
		cfg.APIVersion = "v1"
		cfg.Clusters = map[string]*clientcmdapi.Cluster{
			"cluster1": {Server: "server1"},
		}
		cfg.AuthInfos = map[string]*clientcmdapi.AuthInfo{
			"user1": {Token: "token1"},
		}
		cfg.Contexts = map[string]*clientcmdapi.Context{
			"ctx1": {AuthInfo: "user1", Cluster: "cluster1"},
		}
		clientcmd.WriteToFile(*cfg, fmt.Sprintf(mockKubeConfigPath, i, suffix))
	}
	// cfg := clientcmdapi.NewConfig()
	// cfg.APIVersion = "v1"
	// cfg.Clusters = map[string]*clientcmdapi.Cluster{
	//     "cluster1": {Server: "server1"},
	//     "cluster2": {Server: "server2"},
	//     "cluster3": {Server: "server3"},
	// }
	// cfg.AuthInfos = map[string]*clientcmdapi.AuthInfo{
	//     "user1": {Token: "token1"},
	//     "user2": {Token: "token2"},
	//     "user3": {Token: "token3"},
	// }
	// cfg.Contexts = map[string]*clientcmdapi.Context{
	//     "ctx1": {AuthInfo: "user1", Cluster: "cluster1"},
	//     "ctx2": {AuthInfo: "user2", Cluster: "cluster2"},
	//     "ctx3": {AuthInfo: "user3", Cluster: "cluster3"},
	// }
	// clientcmd.WriteToFile(*cfg, mockKubeConfigPath)
}
