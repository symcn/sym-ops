package clustermanager

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/symcn/sym-ops/pkg/clustermanager/configuration"
	"github.com/symcn/sym-ops/pkg/types"
)

func TestCommon(t *testing.T) {
	t.Run("buildClientCmd with unsupport type", func(t *testing.T) {
		_, err := buildClientCmd(configuration.BuildClusterCfgInfo("", "unknown", "", ""), nil)
		if err == nil {
			t.Error("unsupport ConfigType type must be error")
		}
	})

	t.Run("buildClientCmd with error file", func(t *testing.T) {
		_, err := buildClientCmd(configuration.BuildClusterCfgInfo("", types.KubeConfigTypeFile, "/error/file", ""), nil)
		if err == nil {
			t.Error("error file must be error")
		}
	})

	t.Run("buildClientCmd with error context", func(t *testing.T) {
		_, err := buildClientCmd(configuration.BuildClusterCfgInfo("", types.KubeConfigTypeFile, "", "error-context"), nil)
		if err == nil {
			t.Error("error context must be error")
			return
		}

		home, _ := os.UserHomeDir()
		path := home + "/.kube/config"
		_, err = os.Stat(path)
		if err == nil {
			data, _ := ioutil.ReadFile(path)
			_, err = buildClientCmd(configuration.BuildClusterCfgInfo("", types.KubeConfigTypeRawString, string(data), "error-context"), nil)

			if err == nil {
				t.Error("error context must be error")
				return
			}
		}
	})

	t.Run("buildClientCmd with empty kubeconfig", func(t *testing.T) {
		_, err := buildClientCmd(configuration.BuildClusterCfgInfo("", types.KubeConfigTypeRawString, "", ""), nil)
		if err == nil {
			t.Error("empty rawstring kubeconfig must be error")
		}
	})

	t.Run("healthRequestWithTimeout time less 100ms", func(t *testing.T) {
		_, err := healthRequestWithTimeout(nil, time.Microsecond*1)
		if err == nil {
			t.Error("healthRequestWithTimeout less than 100ms must be error")
		}
	})

	t.Run("healthRequestWithTimeout request", func(t *testing.T) {
		// kubeInterface := fake.NewSimpleClientset()

		// result, err := healthRequestWithTimeout(kubeInterface, time.Second*1)
		// if err != nil {
		//     t.Error(err)
		//     return
		// }
		// if !result {
		//     t.Error("healthRequestWithTimeout request failed")
		// }
	})
}
