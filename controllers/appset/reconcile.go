package appset

import (
	"context"
	"fmt"
	"time"

	"github.com/symcn/api"
	"github.com/symcn/pkg/clustermanager/client"
	"github.com/symcn/pkg/clustermanager/configuration"
	"github.com/symcn/pkg/clustermanager/handler"
	"github.com/symcn/pkg/clustermanager/predicate"
	"github.com/symcn/pkg/clustermanager/workqueue"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	symctx "github.com/symcn/sym-ops/pkg/context"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type master struct {
	currentCli api.MingleClient
	multiCli   api.MultiMingleClient
	stepList   []step
}

// MasterFeature master feature
func MasterFeature(currentCli api.MingleClient, threadiness int, gotInterval time.Duration, server *utils.Server, opt *client.Options) error {
	m := &master{
		currentCli: currentCli,
	}
	m.registryStep()

	// build queue
	qconf := workqueue.NewQueueConfig(m)
	qconf.Name = types.MasterQueueName
	qconf.Threadiness = threadiness
	qconf.GotInterval = gotInterval
	queue, err := workqueue.Completed(qconf).NewQueue()
	if err != nil {
		return err
	}

	// use current client watch appset
	err = currentCli.Watch(&workloadv1beta1.AppSet{},
		queue,
		handler.NewDefaultTransformNamespacedNameEventHandler(),
		predicate.NamespacePredicate(types.FilterNamespaceAppset...),
		&predicateSpec{},
	)
	if err != nil {
		return err
	}
	server.Add(queue)

	// build multi client
	mcc := client.NewMultiClientConfig()
	// build multi cluster configuration manager
	mcc.ClusterCfgManager = configuration.NewClusterCfgManagerWithCM(
		currentCli.GetKubeInterface(),
		types.MultiClusterCfgConfigmapNamespace,
		types.MultiClusterCfgConfigmapLabels,
		types.MultiClusterCfgConfigmapDataKey,
		types.MultiClusterCfgConfigmapStatusKey,
	)
	mcc.RebuildInterval = time.Second * 10
	mcc.Options = opt
	cc, err := client.Complete(mcc)
	if err != nil {
		return err
	}
	multiCli, err := cc.New()
	if err != nil {
		return err
	}

	// registry each cluster watch resource
	multiCli.RegistryBeforAfterHandler(func(ctx context.Context, cli api.MingleClient) error {
		// watch advdeployment
		err := cli.Watch(&workloadv1beta1.AdvDeployment{},
			queue,
			handler.NewDefaultTransformNamespacedNameEventHandler(),
			predicate.NamespacePredicate(types.FilterNamespaceAdvdeployment...),
		)
		if err != nil {
			return fmt.Errorf("cluster %s Watch advdeployment failed: %v", cli.GetClusterCfgInfo().GetName(), err)
		}

		// cache event
		err = cli.SetIndexField(&corev1.Event{}, "type", func(obj rtclient.Object) []string {
			event := obj.(*corev1.Event)
			return []string{event.Type}
		})
		if err != nil {
			return fmt.Errorf("cluster %s SetIndexField corev1.Event failed: %v", cli.GetClusterCfgInfo().GetName(), err)
		}

		return nil
	})
	server.Add(multiCli)
	m.multiCli = multiCli

	return nil
}

func (m *master) registryStep() {
	m.stepList = []step{
		m.stepCheckDeletionTime,
		m.stepAddFinalizer,
		m.stepApplySpec,
		m.stepApplyStatus,
		m.stepDeleteUnuseAdvDeployment,
	}
}

func (m *master) Reconcile(req ktypes.NamespacedName) (api.NeedRequeue, time.Duration, error) {
	app := &workloadv1beta1.AppSet{}
	ctx := symctx.WithValue(context.TODO(), types.ContextKeyStepStop, false)

	err := m.currentCli.Get(req, app)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// maybe some cluster reconnected, will DeleteAll again
			m.deleteAllClusterAdvdeployment(req, nil)
			return api.Done, 0, nil
		}
		return api.Done, 0, fmt.Errorf("Get AppSet %s failed: %v", req, err)
	}

	for _, stepFunc := range m.stepList {
		err = stepFunc(ctx, req, app)
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
