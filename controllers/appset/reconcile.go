package appset

import (
	"context"
	"errors"
	"math/rand"
	"time"

	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	"github.com/symcn/sym-ops/pkg/clustermanager/handler"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/workqueue"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type master struct {
	*rand.Rand
}

// MasterFeature master feature
func MasterFeature(metaCli types.MingleClient) {
	queue, err := workqueue.NewQueue(&master{Rand: rand.New(rand.NewSource(time.Now().Unix()))}, "master", 1, time.Second*1)
	if err != nil {
		klog.Error(err)
		return
	}
	go queue.Run(context.TODO())

	err = metaCli.Watch(&workloadv1beta1.AdvDeployment{}, queue, handler.NewDefaultTransformNamespacedNameEventHandler())
	if err != nil {
		panic(err)
	}
}

func (m *master) Reconcile(req ktypes.NamespacedName) (types.NeedRequeue, time.Duration, error) {
	time.Sleep(time.Millisecond * time.Duration(m.Int31n(100)))
	switch m.Int31n(100) % 4 {
	case 0:
		return types.Requeue, 0, nil
	case 1:
		return types.Done, time.Duration(time.Second * 1), nil
	case 2:
		return types.Done, 0, errors.New("mock error")
	default:
		return types.Done, 0, nil
	}
}
