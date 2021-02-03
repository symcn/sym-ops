package clustermanager

import (
	"context"
	"testing"
	"time"

	"github.com/symcn/sym-ops/pkg/clustermanager/configuration"
	corev1 "k8s.io/api/core/v1"
)

func TestNewMultiClient(t *testing.T) {
	opt := DefaultOptions(nil, 0, 0)
	cli, err := NewMingleClient(&ClientOptions{ClusterCfg: configuration.BuildDefaultClusterCfgInfo("meta")}, opt)
	if err != nil {
		t.Error(err)
		return
	}
	multiOpt := &MultiClientOptions{
		RebuildInterval:             time.Second * 5,
		ClusterConfigurationManager: configuration.NewClusterCfgManagerWithCM(cli.GetKubeInterface(), "sym-admin", map[string]string{"ClusterOwner": "sym-admin"}, "kubeconfig.yaml", "status"),
	}

	multiCli, err := NewMultiMingleClient(multiOpt, opt)
	if err != nil {
		t.Error(err)
		return
	}
	err = multiCli.TriggerSync(&corev1.ConfigMap{})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ch := make(chan struct{}, 0)
	go func() {
		err = multiCli.Start(ctx)
		if err != nil {
			t.Error(err)
		}
		close(ch)
	}()

	syncCh := make(chan struct{}, 0)
	go func() {
		for !multiCli.HasSynced() {
			t.Log("wait sync")
			time.Sleep(time.Millisecond * 100)
		}
		close(syncCh)
	}()

	select {
	case <-ch:
	case <-syncCh:
	}
}
