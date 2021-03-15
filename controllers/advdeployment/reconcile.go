package advdeployment

import (
	"context"
	"fmt"
	"time"

	"github.com/symcn/api"
	"github.com/symcn/pkg/clustermanager"
	"github.com/symcn/pkg/clustermanager/handler"
	"github.com/symcn/pkg/clustermanager/predicate"
	"github.com/symcn/pkg/clustermanager/workqueue"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	symctx "github.com/symcn/sym-ops/pkg/context"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type worker struct {
	currentCli api.MingleClient
	stepList   []step
	conf       *AdvConfig
}

// WorkerFeature worker feature
func WorkerFeature(currentCli api.MingleClient, threadiness int, gotInterval time.Duration, advConf *AdvConfig, server *utils.Server, opt *clustermanager.Options) error {
	if advConf == nil {
		advConf = DefaultAdvConfig()
	}
	w := &worker{
		currentCli: currentCli,
		conf:       advConf,
	}

	w.initialization()
	w.registryStep()

	// build queue
	queue, err := workqueue.NewQueue(w, types.WorkerQueueName, threadiness, gotInterval)
	if err != nil {
		return err
	}

	server.Add(queue)
	return nil
}

func (w *worker) watchResource(queue api.WorkQueue) error {
	// Advdeployment
	err := w.currentCli.Watch(&workloadv1beta1.AdvDeployment{},
		queue,
		handler.NewDefaultTransformNamespacedNameEventHandler(),
		predicate.NamespacePredicate(types.FilterNamespaceAdvdeployment...),
		&predicateSpec{},
	)
	if err != nil {
		return err
	}

	// Deployment
	err = w.currentCli.Watch(&appsv1.Deployment{},
		queue,
		&transformNamespace{},
		predicate.NamespacePredicate(types.FilterNamespaceAdvdeployment...),
		predicate.LabelsKeyPredicate(types.ObserveMustLabelClusterName, types.ObserveMustLabelAppName),
	)
	if err != nil {
		return err
	}

	// StatefulSet
	err = w.currentCli.Watch(&appsv1.StatefulSet{},
		queue,
		&transformNamespace{},
		predicate.NamespacePredicate(types.FilterNamespaceAdvdeployment...),
		predicate.LabelsKeyPredicate(types.ObserveMustLabelClusterName, types.ObserveMustLabelAppName),
	)
	if err != nil {
		return err
	}

	// Job
	err = w.currentCli.Watch(&batchv1.Job{},
		queue,
		&transformNamespace{},
		predicate.NamespacePredicate(types.FilterNamespaceAdvdeployment...),
		predicate.LabelsKeyPredicate(types.ObserveMustLabelClusterName, types.ObserveMustLabelAppName),
	)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) registryStep() {
	w.stepList = []step{
		w.stepCheckDeletionTime,
		w.stepCheckType,
		w.stepApplyResources,
		w.stepRecalculateStatus,
		w.stepUpdateStatus,
	}
}

func (w *worker) Reconcile(req ktypes.NamespacedName) (api.NeedRequeue, time.Duration, error) {
	adv := &workloadv1beta1.AdvDeployment{}
	ctx := symctx.WithValue(context.TODO(), types.ContextKeyStepStop, false)

	err := w.currentCli.Get(req, adv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(3).Infof("Not found Advdeployment %s, skip", req)
			return api.Done, 0, nil
		}
		return api.Done, 0, fmt.Errorf("Get Advdeployment %s failed: %v", req, err)
	}

	for _, stepFunc := range w.stepList {
		err = stepFunc(ctx, req, adv)
		if symctx.GetValueBool(ctx, types.ContextKeyStepStop) {
			// no need exec next step
			break
		}
		// if continue, just print error info
		if err != nil {
			klog.Error(err)
		}
	}

	// if not continue, the err will return
	return api.NeedRequeue(symctx.GetValueBool(ctx, types.ContextKeyNeedRequeue)), time.Duration(symctx.GetValueInt64(ctx, types.ContextKeyRequeueAfter)), err
}
