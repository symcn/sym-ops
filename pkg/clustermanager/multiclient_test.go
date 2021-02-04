package clustermanager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/symcn/sym-ops/pkg/clustermanager/configuration"
	"github.com/symcn/sym-ops/pkg/clustermanager/predicate"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/workqueue"
	corev1 "k8s.io/api/core/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	mockOpt = DefaultOptions(nil, 0, 0)
)

func TestNewMultiClient(t *testing.T) {
	cli, err := NewMingleClient(&ClientOptions{ClusterCfg: configuration.BuildDefaultClusterCfgInfo("meta")}, mockOpt)
	if err != nil {
		t.Error(err)
		return
	}
	multiOpt := &MultiClientOptions{
		RebuildInterval:             time.Second * 5,
		ClusterConfigurationManager: configuration.NewClusterCfgManagerWithCM(cli.GetKubeInterface(), "sym-admin", map[string]string{"ClusterOwner": "sym-admin"}, "kubeconfig.yaml", "status"),
	}

	multiCli, err := NewMultiMingleClient(multiOpt, mockOpt)
	if err != nil {
		t.Error(err)
		return
	}
	err = multiCli.TriggerSync(&corev1.ConfigMap{})
	if err != nil {
		t.Error(err)
		return
	}

	queue, err := workqueue.NewQueue(&reconcile{}, "mockreconcile", 1, time.Second*1)
	if err != nil {
		t.Error(err)
		return
	}
	eventHandler := &mockEventHandler{}
	err = multiCli.Watch(&corev1.Pod{}, queue, eventHandler, predicate.NamespacePredicate("*"))
	if err != nil {
		t.Error(err)
		return
	}

	err = multiCli.Watch(&corev1.ConfigMap{}, queue, eventHandler, predicate.NamespacePredicate("*"))
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

type reconcile struct {
}

func (r *reconcile) Reconcile(req ktypes.NamespacedName) (requeue types.NeedRequeue, after time.Duration, err error) {
	return types.Done, 0, nil
}

type mockEventHandler struct {
}

func (t *mockEventHandler) Create(obj rtclient.Object, queue types.WorkQueue) {
	gvks, b, err := mockOpt.Scheme.ObjectKinds(obj)
	if err != nil {
		fmt.Println(err)
		return
	}
	if b {
		return
	}
	if len(gvks) == 1 {
		fmt.Println(gvks[0].Kind)
	}
}

func (t *mockEventHandler) Update(oldObj, newObj rtclient.Object, queue types.WorkQueue) {
}

func (t *mockEventHandler) Delete(obj rtclient.Object, queue types.WorkQueue) {
}

func (t *mockEventHandler) Generic(obj rtclient.Object, queue types.WorkQueue) {
}
