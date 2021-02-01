package clustermanager

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/symcn/sym-ops/pkg/clustermanager/configuration"
	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	restfake "k8s.io/client-go/rest/fake"
)

var codecs = serializer.NewCodecFactory(scheme.Scheme)

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
		_, err := healthRequestWithTimeout(nil, time.Second*1)
		if err == nil {
			t.Error("healthRequestWithTimeout less than 100ms must be error")
		}
	})

	t.Run("healthRequestWithTimeout client is nil", func(t *testing.T) {
		restCli := &restfake.RESTClient{}
		_, err := healthRequestWithTimeout(restCli, time.Microsecond*1)
		if err == nil {
			t.Error("healthRequestWithTimeout less than 100ms must be error")
		}
	})

	t.Run("healthRequestWithTimeout request not ok", func(t *testing.T) {
		restCli := &restfake.RESTClient{
			Client: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				header := http.Header{}
				header.Set("Content-Type", runtime.ContentTypeJSON)
				return &http.Response{StatusCode: http.StatusOK, Header: header, Body: ioutil.NopCloser(bytes.NewReader([]byte("mock faile resp")))}, nil
			}),
			NegotiatedSerializer: codecs.WithoutConversion(),
			GroupVersion:         schema.GroupVersion{},
		}

		result, err := healthRequestWithTimeout(restCli, time.Second*1)
		if err != nil {
			t.Error(err)
			return
		}
		if result {
			t.Error("healthRequestWithTimeout request succ. return mock failed resp, should failed")
		}
	})

	t.Run("healthRequestWithTimeout request ok", func(t *testing.T) {
		restCli := &restfake.RESTClient{
			Client: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				header := http.Header{}
				header.Set("Content-Type", runtime.ContentTypeJSON)
				return &http.Response{StatusCode: http.StatusOK, Header: header, Body: ioutil.NopCloser(bytes.NewReader([]byte("ok")))}, nil
			}),
			NegotiatedSerializer: codecs.WithoutConversion(),
			GroupVersion:         schema.GroupVersion{},
		}

		result, err := healthRequestWithTimeout(restCli, time.Second*1)
		if err != nil {
			t.Error(err)
			return
		}
		if !result {
			t.Error("healthRequestWithTimeout request failed")
		}
	})
}
