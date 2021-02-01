package clustermanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/symcn/sym-ops/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildClientCmd(cfg types.ClusterCfgInfo, setRestConfigFnList []types.SetKubeRestConfig) (*rest.Config, error) {
	switch cfg.GetKubeConfigType() {
	case types.KubeConfigTypeRawString:
		return buildClientCmdWithRawConfig(cfg.GetKubeConfig(), cfg.GetKubeContext(), setRestConfigFnList)
	case types.KubeConfigTypeFile:
		return buildClientCmdWithFile(cfg.GetKubeConfig(), cfg.GetKubeContext(), setRestConfigFnList)
	default:
		return nil, errors.New("just supoort rawstring and file kubeconfig")
	}
}

func buildClientCmdWithRawConfig(kubeconf string, kubecontext string, setRestConfigFnList []types.SetKubeRestConfig) (*rest.Config, error) {
	if kubeconf == "" {
		return nil, errors.New("kubeconfig is empty")
	}
	apiConfig, err := clientcmd.Load([]byte(kubeconf))
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes API config:%+v", err)
	}

	restcfg, err := clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{CurrentContext: kubecontext}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config from API config:%+v", err)
	}

	for _, fn := range setRestConfigFnList {
		fn(restcfg)
	}
	return restcfg, nil
}

func buildClientCmdWithFile(kubeconf string, kubecontext string, setRestConfigFnList []types.SetKubeRestConfig) (*rest.Config, error) {
	if kubeconf != "" {
		info, err := os.Stat(kubeconf)
		if err != nil || info.Size() == 0 {
			return nil, fmt.Errorf("file %s not exists or empty", kubeconf)
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconf
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: kubecontext,
	}

	restcfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	for _, fn := range setRestConfigFnList {
		fn(restcfg)
	}
	return restcfg, nil
}

func buildKubeInterface(restcfg *rest.Config) (kubernetes.Interface, error) {
	clientSet, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func healthRequestWithTimeout(restCli rest.Interface, timeout time.Duration) (bool, error) {
	if restCli == nil {
		return false, errors.New("health request rest client is nil")
	}

	// Always return false, when the timeout too small, so must large than 100ms
	if timeout < minExectimeout {
		return false, errors.New("health request timeout must more than 100ms")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	body, err := restCli.Get().AbsPath("/healthz").Do(ctx).Raw()
	if err != nil {
		return false, err
	}
	return strings.EqualFold(string(body), "ok"), nil
}
