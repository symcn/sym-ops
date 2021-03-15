package appset

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/symcn/api"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	symctx "github.com/symcn/sym-ops/pkg/context"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *master) stepCheckDeletionTime(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error {
	if app.ObjectMeta.DeletionTimestamp.IsZero() {
		return nil
	}

	// if do this, next step not continue
	symctx.WithValue(ctx, types.ContextKeyStepStop, true)
	err := m.deleteAllClusterAdvdeployment(req, app)
	if err != nil {
		// just print error and set after requeueAfterTimeError
		klog.Error(err)
		symctx.WithValue(ctx, types.ContextKeyRequeueAfter, requeueAfterTimeError)
	}
	return nil
}

func (m *master) deleteAllClusterAdvdeployment(req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error {
	isChanged, errMsg := m.concurrentExecStepForEachCluster(req, app, m.deleteAdvdeploymentWithClusterClient)
	if len(errMsg) > 0 {
		// just print error, if all cluster delete failure, will remove finalizer,
		// when the cluster reconnected, will re-reconcile, now the Appset not found, so just delete Advdeployment.
		klog.Errorf("Delete all Advdeployment has error: %v", strings.Join(errMsg, errorSep))
	}

	if isChanged || app == nil {
		// maybe the Appset already delete
		return nil
	}

	if len(app.ObjectMeta.Finalizers) == 0 {
		klog.V(4).Infof("AppSet %s finalizers is empty", req)
		return nil
	}

	klog.V(4).Infof("Appset %s delete all Advdeployment success, remove finalizer now.", req)
	app.ObjectMeta.Finalizers = utils.RemoveSliceString(app.ObjectMeta.Finalizers, types.FinalizersStr)
	err := m.currentCli.Update(app)
	if err != nil {
		return fmt.Errorf("Appset %s remove finalizers failed: %v", req, err)
	}
	return nil
}

func (m *master) deleteAdvdeploymentWithClusterClient(cli api.MingleClient, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) (bool, error) {
	err := cli.Delete(&workloadv1beta1.AdvDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	})
	if err == nil {
		klog.V(4).Infof("Delete cluster %s Advdeployment %s successfully.", cli.GetClusterCfgInfo().GetName(), req)
		return true, nil
	}

	if apierrors.IsNotFound(err) {
		klog.Warningf("Delete cluster %s Advdeployment %s failed, not found", cli.GetClusterCfgInfo().GetName(), req)
		return false, nil
	}

	return false, fmt.Errorf("Delete cluster %s Advdeployment %s failed: %v", cli.GetClusterCfgInfo().GetName(), req, err)
}

func (m *master) stepAddFinalizer(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error {
	if utils.SliceContainsString(app.ObjectMeta.Finalizers, types.FinalizersStr) {
		klog.V(5).Infof("Appset %s has finalizer, skip", req)
		return nil
	}

	symctx.WithValue(ctx, types.ContextKeyStepStop, true)

	if app.ObjectMeta.Finalizers == nil {
		app.ObjectMeta.Finalizers = []string{}
	}
	app.ObjectMeta.Finalizers = append(app.ObjectMeta.Finalizers, types.FinalizersStr)
	err := m.currentCli.Update(app)
	if err != nil {
		return fmt.Errorf("Appset %s set finalizers failed: %v", req, err)
	}
	return nil
}

func (m *master) stepApplySpec(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error {
	f := func(deployClusterSpec *workloadv1beta1.TargetCluster, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) (bool, error) {
		cli, err := m.multiCli.GetConnectedWithName(deployClusterSpec.Name)
		if err != nil {
			return false, err
		}
		obj := buildAdvdeploymentWithApp(app, deployClusterSpec)
		return m.applyAdvdeployment(cli, req, obj)
	}

	isChanged, errs := m.concurrentExecStepForAppsetSpecifyCluster(req, app, f)
	if len(errs) > 0 {
		klog.Errorf("Apply Appset %s spec failed: %s", req, strings.Join(errs, errorSep))
	}

	if isChanged {
		symctx.WithValue(ctx, types.ContextKeyStepStop, true)
		symctx.WithValue(ctx, types.ContextKeyRequeueAfter, requeueAfterTimeGrace)
	}
	return nil
}

func (m *master) applyAdvdeployment(cli api.MingleClient, req ktypes.NamespacedName, new *workloadv1beta1.AdvDeployment) (isChanged bool, err error) {
	old := &workloadv1beta1.AdvDeployment{}
	err = cli.Get(req, old)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// not found, create it
			new.Status.AggrStatus.Status = workloadv1beta1.AppStatusInstalling
			err = cli.Create(new)
			if err != nil {
				return false, fmt.Errorf("Create %s cluster Advdeployment %s failed: %v", cli.GetClusterCfgInfo().GetName(), req, err)
			}
			klog.V(4).Infof("Create %s cluster Advdeployment %s successfully", cli.GetClusterCfgInfo().GetName(), req)
			return true, nil
		}
		return false, fmt.Errorf("Get %s Advdeployment %s failed: %+v", cli.GetClusterCfgInfo().GetName(), req, err)
	}

	if !isAdvdeploymentDifferent(old, new) {
		// spec not modify
		return false, nil
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		time := metav1.Now()
		new.Spec.DeepCopyInto(&old.Spec)
		new.Finalizers = old.Finalizers
		new.Labels = old.Labels
		new.Annotations = old.Annotations
		new.Status.LastUpdateTime = &time

		updateErr := cli.Update(new)
		if updateErr == nil {
			klog.V(4).Infof("Update %s cluster Advdeployment %s successfully", cli.GetClusterCfgInfo().GetName(), req)
			return nil
		}

		getErr := cli.Get(req, old)
		if getErr != nil {
			klog.Errorf("Re-get %s Advdeployment %s failed: %+v", cli.GetClusterCfgInfo().GetName(), req, getErr)
		}

		if !isAdvdeploymentDifferent(old, new) {
			// same spec not need update
			// ... is invalid: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update
			return nil
		}
		return updateErr
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *master) stepApplyStatus(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error {
	as := m.buildAppsetStatus(req, app)
	if app.Status.ObservedGeneration == as.ObservedGeneration && equality.Semantic.DeepEqual(app.Status.AggrStatus, as.AggrStatus) {
		klog.V(3).Infof("Appset %s status unchanged", req)
		return nil
	}

	if as.AggrStatus.Status == workloadv1beta1.AppStatusRuning && app.Status.AggrStatus.Status != workloadv1beta1.AppStatusRuning {
		// just record one event
		// TODO record event
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		as.AggrStatus.DeepCopyInto(&app.Status.AggrStatus)
		app.Status.ObservedGeneration = as.ObservedGeneration
		t := metav1.Now()
		app.Status.LastUpdateTime = &t

		updateErr := m.currentCli.StatusUpdate(app)
		if updateErr == nil {
			klog.V(3).Infof("Update Appset %s status %s successfully", req, app.Status.AggrStatus.Status)
			return nil
		}

		getErr := m.currentCli.Get(req, app)
		if getErr != nil {
			klog.Errorf("Re-get Appset %s failed: %v", req, getErr)
			return getErr
		}
		if app.Status.ObservedGeneration == as.ObservedGeneration && equality.Semantic.DeepEqual(app.Status.AggrStatus, as.AggrStatus) {
			// same status not need update
			klog.V(3).Infof("Re-get Appset compare status is equal.")
			return nil
		}
		return updateErr
	})
	if err != nil {
		symctx.WithValue(ctx, types.ContextKeyStepStop, true)
		return fmt.Errorf("Update Appset %s status failed: %v", req, err)
	}

	symctx.WithValue(ctx, types.ContextKeyAppsetStatus, as.AggrStatus.Status)
	return nil
}

func (m *master) buildAppsetStatus(req ktypes.NamespacedName, app *workloadv1beta1.AppSet) *workloadv1beta1.AppSetStatus {
	as := &workloadv1beta1.AppSetStatus{
		AggrStatus: workloadv1beta1.AggrAppSetStatus{
			Pods:       []*workloadv1beta1.Pod{},
			Clusters:   []*workloadv1beta1.ClusterAppActual{},
			WarnEvents: []*workloadv1beta1.Event{},
		},
	}
	nsAdvs := m.getAllClusterComplexAdvdeployment(req, app)
	var (
		changeObserved = true
		finalStatus    = workloadv1beta1.AppStatusRuning
	)
	for _, nsAdv := range nsAdvs {
		as.AggrStatus.Version = mergeVersion(as.AggrStatus.Version, nsAdv.Adv.Status.AggrStatus.Version)
		as.AggrStatus.Clusters = append(as.AggrStatus.Clusters, &workloadv1beta1.ClusterAppActual{
			Name:        nsAdv.ClusterName,
			Desired:     nsAdv.Adv.Status.AggrStatus.Desired,
			Available:   nsAdv.Adv.Status.AggrStatus.Available,
			UnAvailable: nsAdv.Adv.Status.AggrStatus.UnAvailable,
			PodSets:     nsAdv.Adv.Status.AggrStatus.PodSets,
		})

		as.AggrStatus.Desired += nsAdv.Adv.Status.AggrStatus.Desired
		as.AggrStatus.Available += nsAdv.Adv.Status.AggrStatus.Available
		as.AggrStatus.UnAvailable += nsAdv.Adv.Status.AggrStatus.UnAvailable

		if changeObserved {
			changeObserved = nsAdv.Adv.ObjectMeta.Generation == nsAdv.Adv.Status.ObservedGeneration
		}

		if nsAdv.Adv.ObjectMeta.Generation != nsAdv.Adv.Status.ObservedGeneration || nsAdv.Adv.Status.AggrStatus.Status != workloadv1beta1.AppStatusRuning {

			klog.V(5).Infof("Cluster %s advdeployment %s status is %s meta generation:%d, observedGeneration:%d",
				nsAdv.ClusterName,
				req.String(),
				nsAdv.Adv.Status.AggrStatus.Status,
				nsAdv.Adv.ObjectMeta.Generation,
				nsAdv.Adv.Status.ObservedGeneration)

			finalStatus = workloadv1beta1.AppStatusInstalling
		}
	}
	var replicas int32
	if app.Spec.Replicas != nil {
		as.AggrStatus.Desired = *app.Spec.Replicas
		replicas = *app.Spec.Replicas
	} else {
		replicas = as.AggrStatus.Desired
	}

	// final status aggregate
	if finalStatus == workloadv1beta1.AppStatusRuning && as.AggrStatus.Available == replicas && as.AggrStatus.UnAvailable == 0 {
		as.AggrStatus.Status = workloadv1beta1.AppStatusRuning
	} else {
		as.AggrStatus.Status = workloadv1beta1.AppStatusInstalling
		as.AggrStatus.WarnEvents = m.getAllClusterWorkloadEnvet(req, app)
	}
	klog.V(5).Infof("Appset %s status:%s, desired:%d, available:%d, replicas:%d, finalStatus:%s", req, finalStatus, as.AggrStatus.Desired, as.AggrStatus.Available, *app.Spec.Replicas, as.AggrStatus.Status)

	if changeObserved {
		as.ObservedGeneration = app.ObjectMeta.Generation
	} else {
		// minus 1 to case previous fetch cached adv which may cause dirty data in this ObservedGeneration
		as.ObservedGeneration = app.ObjectMeta.Generation - 1
	}

	return as
}

func (m *master) getAllClusterComplexAdvdeployment(req ktypes.NamespacedName, app *workloadv1beta1.AppSet) []*complexAdvdeployment {
	var (
		complexAdvdeploymentList = []*complexAdvdeployment{}
		complexAdvdeploymentCh   = make(chan *complexAdvdeployment, 0)
		done                     = make(chan struct{}, 0)
	)
	go func() {
		for ca := range complexAdvdeploymentCh {
			complexAdvdeploymentList = append(complexAdvdeploymentList, ca)
		}
		close(done)
	}()

	f := func(deployClusterSpec *workloadv1beta1.TargetCluster, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) (bool, error) {
		cli, err := m.multiCli.GetConnectedWithName(deployClusterSpec.Name)
		if err != nil {
			return false, fmt.Errorf("Get cluster %s connection failed: %v", deployClusterSpec.Name, err)
		}

		// get advdeployment
		obj := &workloadv1beta1.AdvDeployment{}
		err = cli.Get(req, obj)
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.Warningf("Get cluster %s advdeployment %s not found, maybe the cache not sync", deployClusterSpec.Name, req)
				return false, nil
			}
			return false, fmt.Errorf("Get cluster %s advdeployment %s failed: %v", deployClusterSpec.Name, req, err)
		}
		complexAdvdeploymentCh <- &complexAdvdeployment{
			ClusterName: deployClusterSpec.Name,
			Adv:         obj,
		}

		return false, nil
	}

	_, errs := m.concurrentExecStepForAppsetSpecifyCluster(req, app, f)
	if len(errs) > 0 {
		klog.Errorf("Get all cluster complexAdvdeployment have some err: %s", strings.Join(errs, errorSep))
	}
	close(complexAdvdeploymentCh)
	<-done

	return complexAdvdeploymentList
}

func (m *master) getAllClusterWorkloadEnvet(req ktypes.NamespacedName, app *workloadv1beta1.AppSet) []*workloadv1beta1.Event {
	var (
		workloadEventList = []*workloadv1beta1.Event{}
		eventList         = []*corev1.Event{}
		eventCh           = make(chan *corev1.Event, 0)
		done              = make(chan struct{}, 0)
		eventOption       = &client.ListOptions{
			Namespace:     req.Namespace,
			FieldSelector: fields.Set{"type": corev1.EventTypeWarning}.AsSelector(),
		}
	)
	go func() {
		for e := range eventCh {
			eventList = append(eventList, e)
		}
		eventList = removeDuplicatesEvent(eventList)
		workloadEventList = transformWorkloadEvent(eventList)
		sort.Slice(workloadEventList, func(i int, j int) bool {
			return workloadEventList[i].Name > workloadEventList[j].Name
		})
		close(done)
	}()

	f := func(deployClusterSpec *workloadv1beta1.TargetCluster, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) (bool, error) {
		cli, err := m.multiCli.GetConnectedWithName(deployClusterSpec.Name)
		if err != nil {
			return false, fmt.Errorf("Get cluster %s connection failed: %v", deployClusterSpec.Name, err)
		}

		// get events
		events := &corev1.EventList{}
		err = cli.List(events, eventOption)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("Get cluster %s events list failed: %v", deployClusterSpec.Name, err)
		}

		for _, event := range events.Items {
			if event.InvolvedObject.Kind == "AdvDeployment" && event.InvolvedObject.Name == req.Name {
				eventCh <- &event
				continue
			}

			if event.InvolvedObject.Kind == "Deployment" ||
				event.InvolvedObject.Kind == "StatefulSet" ||
				event.InvolvedObject.Kind == "Pod" ||
				event.InvolvedObject.Kind == "Job" {

				if checkEventLabel(event.InvolvedObject.Name, req.Name) {
					eventCh <- &event
				}
			}
		}

		return false, nil
	}

	_, errs := m.concurrentExecStepForAppsetSpecifyCluster(req, app, f)
	if len(errs) > 0 {
		klog.Errorf("Get all cluster workloadEvent have some err: %s", strings.Join(errs, errorSep))
	}
	close(eventCh)
	<-done

	return workloadEventList
}

func (m *master) stepDeleteUnuseAdvDeployment(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error {
	status := symctx.GetValueString(ctx, types.ContextKeyAppsetStatus)
	if status != string(workloadv1beta1.AppStatusRuning) {
		if v, ok := app.Annotations[deleteUnexpectWaitAllReadyLable]; !ok || !strings.EqualFold(v, "true") {
			// status is not ready, need judge zone
			unexpectList := m.getUnexpectAdvdeploymentClusterListSync(req, app)
			m.deleteUnexpectClusterAdv(unexpectList, req)
			return nil
		}
		// set deleteUnexpectWaitAllReadyLable or not set will wait all ready
		return nil
	}

	// all ready
	unexpectList := m.getUnexpectAdvdeploymentClusterList(req, app)
	m.deleteUnexpectClusterAdv(unexpectList, req)
	return nil
}

func (m *master) deleteUnexpectClusterAdv(unexpectClusterList []string, req ktypes.NamespacedName) {
	if len(unexpectClusterList) < 1 {
		return
	}

	for _, unexpectClusterName := range unexpectClusterList {
		cli, err := m.multiCli.GetConnectedWithName(unexpectClusterName)
		if err != nil {
			klog.Errorf("Delete unexpect cluster %s Advdeployment get connection failed: %v", unexpectClusterName, err)
			continue
		}

		err = cli.Delete(&workloadv1beta1.AdvDeployment{ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace}})
		if err == nil {
			klog.V(4).Info("Delete unexpect cluster %s Advdeployment %s successfully", unexpectClusterName, req)
			continue
		}
		klog.Errorf("Delete unexpect cluster %s Advdeployment %s failed: %v", unexpectClusterName, req, err)
	}
	return
}
