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
	cli, err := NewMingleClient(configuration.BuildDefaultClusterCfgInfo("meta"), mockOpt)
	if err != nil {
		t.Error(err)
		return
	}
	clusterCfgManager := configuration.NewClusterCfgManagerWithCM(cli.GetKubeInterface(), "sym-admin", map[string]string{"ClusterOwner": "sym-admin"}, "kubeconfig.yaml", "status")

	multiCli, err := NewMultiMingleClient(clusterCfgManager, time.Second*5, mockOpt)
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
	// eventHandler := &mockEventHandler{}
	// err = multiCli.Watch(&corev1.Pod{}, queue, eventHandler, predicate.NamespacePredicate("*"))
	// if err != nil {
	//     t.Error(err)
	//     return
	// }

	// err = cli.Watch(&corev1.ConfigMap{}, queue, eventHandler, predicate.NamespacePredicate("*"))
	// if err != nil {
	//     t.Error(err)
	//     return
	// }
	multiCli.RegistryBeforAfterHandler(func(cli types.MingleClient) error {
		eventHandler := &mockEventHandler{}
		err := cli.Watch(&corev1.Pod{}, queue, eventHandler, predicate.NamespacePredicate("*"))
		if err != nil {
			t.Error(err)
			return err
		}

		err = cli.Watch(&corev1.ConfigMap{}, queue, eventHandler, predicate.NamespacePredicate("*"))
		if err != nil {
			t.Error(err)
			return err
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ch := make(chan struct{}, 0)
	go func() {
		err = queue.Run(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

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
	fmt.Println(req.String())
	return types.Done, 0, nil
}

type mockEventHandler struct {
}

func (t *mockEventHandler) Create(obj rtclient.Object, queue types.WorkQueue) {
	// gvks, b, err := mockOpt.Scheme.ObjectKinds(obj)
	// if err != nil {
	//     fmt.Println(err)
	//     return
	// }
	// if b {
	//     return
	// }
	// if len(gvks) == 1 {
	//     fmt.Println(gvks[0].Kind)
	// }
	queue.Add(ktypes.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()})
}

func (t *mockEventHandler) Update(oldObj, newObj rtclient.Object, queue types.WorkQueue) {
}

func (t *mockEventHandler) Delete(obj rtclient.Object, queue types.WorkQueue) {
}

func (t *mockEventHandler) Generic(obj rtclient.Object, queue types.WorkQueue) {
}
