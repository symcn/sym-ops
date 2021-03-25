package advdeployment

import (
	"context"
	"errors"
	"fmt"
	"sort"

	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	symctx "github.com/symcn/sym-ops/pkg/context"
	"github.com/symcn/sym-ops/pkg/helm"
	"github.com/symcn/sym-ops/pkg/resource"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (w *worker) stepCheckDeletionTime(ctx context.Context, req ktypes.NamespacedName, adv *workloadv1beta1.AdvDeployment) error {
	if adv.ObjectMeta.DeletionTimestamp.IsZero() {
		return nil
	}

	symctx.WithValue(ctx, types.ContextKeyStepStop, true)
	return nil
}

func (w *worker) stepCheckType(ctx context.Context, req ktypes.NamespacedName, adv *workloadv1beta1.AdvDeployment) error {
	var err error
	defer func() {
		if err != nil {
			symctx.WithValue(ctx, types.ContextKeyStepStop, true)
			symctx.WithValue(ctx, types.ContextKeyRequeueAfter, requeueAfterTime)
			w.currentCli.Eventf(adv, corev1.EventTypeWarning, "Check failed: %s", err.Error())
		}
	}()

	if adv.Spec.PodSpec.DeployType != "helm" {
		err = fmt.Errorf("Advdeployment %s not supported type %s", req, adv.Spec.PodSpec.DeployType)
		return err
	}
	if adv.Spec.PodSpec.Chart == nil {
		err = fmt.Errorf("Advdeployment %s chart is nil", req)
		return err
	}
	if adv.Spec.PodSpec.Chart.CharURL == nil && adv.Spec.PodSpec.Chart.RawChart == nil {
		err = fmt.Errorf("Advdeployment %s char url and raw chart not both nil", req)
		return err
	}
	return nil
}

func (w *worker) stepApplyResources(ctx context.Context, req ktypes.NamespacedName, adv *workloadv1beta1.AdvDeployment) error {
	var err error
	defer func() {
		if err != nil {
			symctx.WithValue(ctx, types.ContextKeyStepStop, true)
			symctx.WithValue(ctx, types.ContextKeyRequeueAfter, requeueAfterTime)
			w.currentCli.Eventf(adv, corev1.EventTypeWarning, "Apply resource failed: %s", err.Error())
		}
	}()

	var (
		objects []helm.K8sObject
	)
	for _, podSet := range adv.Spec.Topology.PodSets {
		_, _, rawChart := getCharInfo(podSet, adv)
		objs, err := helm.RenderTemplate(rawChart, podSet.Name, adv.Namespace, podSet.RawValues)
		if err != nil {
			return err
		}
		objects = append(objects, objs...)
	}

	ownerRes := []string{}
	isHpaEnable := getHpaSpecEnable(adv.Annotations)
	var (
		changed  bool
		change   int
		rtobj    rtclient.Object
		opt      resource.Option
		replicas int32
	)
	for _, obj := range objects {
		yaml := obj.YAML2String()
		klog.V(5).Infof("%s %s/%s yaml: %s", obj.GroupKind().Kind, obj.GetNamespace(), obj.GetName(), yaml)

		covert, ok := convertFactory[obj.GroupKind().Kind]
		if !ok {
			return fmt.Errorf("Apply resource %s %s/%s unsupport type", obj.GroupKind().Kind, obj.GetNamespace(), obj.GetName())
		}
		rtobj, opt, replicas, err = covert(obj.UnstructuredObject(), isHpaEnable)
		if err != nil {
			return err
		}
		ownerRes = append(ownerRes, getFormattedName(obj.GroupKind().Kind, rtobj))
		changed, err = resource.Reconcile(ctx, w.currentCli, rtobj, opt)
		if err != nil {
			symctx.WithValue(ctx, types.ContextKeyStepStop, true)
			return fmt.Errorf("Apply resource failed: %v", err)
		}
		if obj.GroupKind().Kind == types.DeploymentKind || obj.GroupKind().Kind == types.StatefulSetKind {
			err = w.applyHorizontalPodAutoscaler(ctx, adv, obj, types.HorizontalAPIVersion, replicas)
			if err != nil {
				klog.Error(err)
			}
		}
		if changed {
			change++
		}
	}
	if change > 0 {
		symctx.WithValue(ctx, types.ContextKeyStepStop, true)
		symctx.WithValue(ctx, types.ContextKeyRequeueAfter, requeueAfterTime)
	}
	symctx.WithValue(ctx, types.ContextKeyAdvdeploymentOwnerRes, ownerRes)
	return nil
}

func (w *worker) applyHorizontalPodAutoscaler(ctx context.Context, adv *workloadv1beta1.AdvDeployment, obj helm.K8sObject, apiVersion string, currentReplicas int32) error {
	enable := getHpaSpecEnable(adv.Annotations)
	if !enable || currentReplicas == 0 {
		klog.V(5).Infof("Hpa not enable or object %s/%s replicas is zero", obj.GetNamespace(), obj.GetName())
		hpa := &v2beta2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		}
		resource.Reconcile(ctx, w.currentCli, hpa, resource.Option{DesiredState: resource.DesiredStateAbsent})
		return nil
	}

	metrics := parseMetrics(adv.Annotations, obj.GetName())
	if len(metrics) == 0 {
		metrics = w.getDefaultMetric()
	}

	hpa := &v2beta2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Labels: map[string]string{
				"app":                        adv.Name,
				"app.kubernetes.io/instance": obj.GetName(),
			},
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2beta2.CrossVersionObjectReference{
				APIVersion: apiVersion,
				Kind:       obj.GroupKind().Kind,
				Name:       obj.GetName(),
			},
			Metrics:     metrics,
			MinReplicas: &currentReplicas,
			MaxReplicas: currentReplicas * 2,
		},
	}
	if err := controllerutil.SetControllerReference(adv, hpa, types.Scheme); err != nil {
		klog.Errorf("SetControllerReference failed: %v", err)
	}
	klog.V(5).Infof("Start apply hpa name %s minReplicas %d maxReplicas %d", hpa.Name, *hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas)

	return w.applyResourceForHpa(hpa)
}

func (w *worker) applyResourceForHpa(desired *v2beta2.HorizontalPodAutoscaler) error {
	key := ktypes.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}
	current := &v2beta2.HorizontalPodAutoscaler{}
	err := w.currentCli.Get(key, current)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = w.currentCli.Create(desired)
			if err != nil {
				return fmt.Errorf("create hpa %s failed: %v", key, err)
			}
			klog.Infof("Create hpa %s successfully", key)
			return nil
		}
		return fmt.Errorf("get hpa %s failed: %v", key, err)
	}

	if equalHpa(current, desired) {
		return nil
	}

	metaAccessor := meta.NewAccessor()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentResourceVersion, err := metaAccessor.ResourceVersion(current)
		if err != nil {
			return fmt.Errorf("get hpa %s resource version failed: %v", key, err)
		}
		metaAccessor.SetResourceVersion(desired, currentResourceVersion)

		updateErr := w.currentCli.Update(desired)
		if updateErr == nil {
			klog.V(4).Infof("Update hpa %s successfully", key)
			return nil
		}

		getErr := w.currentCli.Get(key, current)
		if getErr != nil {
			klog.Errorf("Get hpa %s failed: %v", key, getErr)
		}
		return updateErr
	})
}

func (w *worker) stepRecalculateStatus(ctx context.Context, key ktypes.NamespacedName, adv *workloadv1beta1.AdvDeployment) error {
	w.checkServiceOwner(adv)

	var owners []string
	if o, ok := symctx.GetValue(ctx, types.ContextKeyAdvdeploymentOwnerRes).([]string); ok {
		owners = o
	}

	// deployment check
	deploys, err := w.getDeployListByLabels(adv)
	if err != nil {
		return err
	}
	if len(deploys) > 1 {
		unusedObjects, status, updatedReplicas, generationEqual := w.loopDeploys(adv, deploys, owners)
		w.dealAggreStatus(ctx, status, generationEqual, updatedReplicas, unusedObjects)
		return nil
	}

	// statefulset check
	statefulsets, err := w.getStatefulSetListByLabels(adv)
	if err != nil {
		return err
	}
	if len(statefulsets) > 1 {
		unusedObjects, status, updatedReplicas, generationEqual := w.loopStatefulSet(adv, statefulsets, owners)
		w.dealAggreStatus(ctx, status, generationEqual, updatedReplicas, unusedObjects)
		return nil
	}

	// job check
	jobs, err := w.getJobListByLabels(adv)
	if err != nil {
		return err
	}
	if len(statefulsets) > 1 {
		unusedObjects, status, updatedReplicas, generationEqual := w.loopJob(adv, jobs, owners)
		w.dealAggreStatus(ctx, status, generationEqual, updatedReplicas, unusedObjects)
		return nil
	}

	return nil
}

func (w *worker) checkServiceOwner(adv *workloadv1beta1.AdvDeployment) {
	svcList := &corev1.ServiceList{}
	err := w.currentCli.List(svcList, &rtclient.ListOptions{
		Namespace:     adv.Namespace,
		LabelSelector: labels.Set{types.ObserveMustLabelAppName: adv.Name + types.ServiceNameSuffix}.AsSelector(),
	})
	if err != nil {
		klog.Errorf("get service list %s/%s failed: %v", adv.Namespace, adv.Name, err)
		return
	}

	if len(svcList.Items) != 1 {
		klog.Errorf("service list %s/%s have more or less len %d", adv.Namespace, adv.Name, len(svcList.Items))
		return
	}

	w.setControllerReference(adv, &svcList.Items[0])
}

func (w *worker) setControllerReference(adv *workloadv1beta1.AdvDeployment, obj rtclient.Object) {
	if metav1.IsControlledBy(obj, adv) {
		return
	}
	err := controllerutil.SetControllerReference(adv, obj, types.Scheme)
	if err != nil {
		klog.Errorf("SetControllerReference %s %s/%s owner failed: %v", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName(), err)
		return
	}
	err = w.currentCli.Update(obj)
	if err != nil {
		klog.Errorf("Update %s %s/%s owner failed: %v", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName(), err)
		return
	}
	klog.V(4).Infof("Update %s %s/%s owner successfully", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetNamespace(), obj.GetName())
	return
}

func (w *worker) getDeployListByLabels(adv *workloadv1beta1.AdvDeployment) ([]appsv1.Deployment, error) {
	deployList := &appsv1.DeploymentList{}
	err := w.currentCli.List(deployList, &rtclient.ListOptions{
		Namespace:     adv.Namespace,
		LabelSelector: labels.Set{types.ObserveMustLabelAppName: adv.Name}.AsSelector(),
	})
	if err != nil {
		return nil, fmt.Errorf("get deployment list %s/%s failed: %v", adv.Namespace, adv.Name, err)
	}

	return deployList.Items, nil
}

func (w *worker) getStatefulSetListByLabels(adv *workloadv1beta1.AdvDeployment) ([]appsv1.StatefulSet, error) {
	statefulSetList := &appsv1.StatefulSetList{}
	err := w.currentCli.List(statefulSetList, &rtclient.ListOptions{
		Namespace:     adv.Namespace,
		LabelSelector: labels.Set{types.ObserveMustLabelAppName: adv.Name}.AsSelector(),
	})
	if err != nil {
		return nil, fmt.Errorf("get statefulset list %s/%s failed: %v", adv.Namespace, adv.Name, err)
	}

	return statefulSetList.Items, nil
}

func (w *worker) getJobListByLabels(adv *workloadv1beta1.AdvDeployment) ([]batchv1.Job, error) {
	jobList := &batchv1.JobList{}
	err := w.currentCli.List(jobList, &rtclient.ListOptions{
		Namespace:     adv.Namespace,
		LabelSelector: labels.Set{types.ObserveMustLabelAppName: adv.Name}.AsSelector(),
	})
	if err != nil {
		return nil, fmt.Errorf("get job list %s/%s failed: %v", adv.Namespace, adv.Name, err)
	}

	return jobList.Items, nil
}

func (w *worker) loopDeploys(adv *workloadv1beta1.AdvDeployment, deploys []appsv1.Deployment, owners []string) ([]rtclient.Object, *workloadv1beta1.AdvDeploymentAggrStatus, int32, bool) {
	var (
		unusedObjects   = []rtclient.Object{}
		status          = &workloadv1beta1.AdvDeploymentAggrStatus{}
		updatedReplicas int32
		generationEqual = true
	)
	for _, deploy := range deploys {
		w.setControllerReference(adv, &deploy)

		if isUnunseObject(types.DeploymentKind, &deploy, owners) {
			unusedObjects = append(unusedObjects, &deploy)
			continue
		}

		// build podSetStatus
		podSetStatus := &workloadv1beta1.PodSetStatusInfo{}
		podSetStatus.Name = deploy.Name
		podSetStatus.Version = utils.GetPodContainerImageVersion(adv.Name, &deploy.Spec.Template.Spec)
		podSetStatus.Available = deploy.Status.AvailableReplicas
		podSetStatus.Desired = *deploy.Spec.Replicas
		podSetStatus.UnAvailable = deploy.Status.UnavailableReplicas
		podSetStatus.Update = &deploy.Status.UpdatedReplicas
		podSetStatus.Current = &deploy.Status.Replicas
		podSetStatus.Ready = &deploy.Status.ReadyReplicas

		status.PodSets = append(status.PodSets, podSetStatus)

		// count status
		status.Available += podSetStatus.Available
		status.Desired += podSetStatus.Desired
		status.UnAvailable += podSetStatus.UnAvailable

		// add update replicas
		updatedReplicas += deploy.Status.UpdatedReplicas

		if deploy.Status.ObservedGeneration != deploy.ObjectMeta.Generation {
			// !import must all deploy both equal should equal
			generationEqual = false
		}
	}
	return unusedObjects, status, updatedReplicas, generationEqual
}

func (w *worker) loopStatefulSet(adv *workloadv1beta1.AdvDeployment, statefulSets []appsv1.StatefulSet, owners []string) ([]rtclient.Object, *workloadv1beta1.AdvDeploymentAggrStatus, int32, bool) {
	var (
		unusedObjects   = []rtclient.Object{}
		status          = &workloadv1beta1.AdvDeploymentAggrStatus{}
		updatedReplicas int32
		generationEqual = true
	)
	for _, statefulset := range statefulSets {
		w.setControllerReference(adv, &statefulset)

		if isUnunseObject(types.DeploymentKind, &statefulset, owners) {
			unusedObjects = append(unusedObjects, &statefulset)
			continue
		}

		// build podSetStatus
		podSetStatus := &workloadv1beta1.PodSetStatusInfo{}
		podSetStatus.Name = statefulset.Name
		podSetStatus.Version = utils.GetPodContainerImageVersion(adv.Name, &statefulset.Spec.Template.Spec)
		podSetStatus.Available = statefulset.Status.ReadyReplicas
		podSetStatus.Desired = *statefulset.Spec.Replicas
		podSetStatus.Update = &statefulset.Status.UpdatedReplicas
		podSetStatus.Current = &statefulset.Status.Replicas
		podSetStatus.Ready = &statefulset.Status.ReadyReplicas

		status.PodSets = append(status.PodSets, podSetStatus)

		// count status
		status.Available += podSetStatus.Available
		status.Desired += podSetStatus.Desired

		// add update replicas
		updatedReplicas += statefulset.Status.UpdatedReplicas

		if statefulset.Status.ObservedGeneration != statefulset.ObjectMeta.Generation {
			generationEqual = false
		}
	}
	return unusedObjects, status, updatedReplicas, generationEqual
}

func (w *worker) loopJob(adv *workloadv1beta1.AdvDeployment, jobs []batchv1.Job, owners []string) ([]rtclient.Object, *workloadv1beta1.AdvDeploymentAggrStatus, int32, bool) {
	var (
		unusedObjects   = []rtclient.Object{}
		status          = &workloadv1beta1.AdvDeploymentAggrStatus{}
		updatedReplicas int32
		generationEqual = true
	)
	for _, job := range jobs {
		w.setControllerReference(adv, &job)

		if isUnunseObject(types.DeploymentKind, &job, owners) {
			unusedObjects = append(unusedObjects, &job)
			continue
		}

		// build podSetStatus
		podSetStatus := &workloadv1beta1.PodSetStatusInfo{}
		podSetStatus.Name = job.Name
		podSetStatus.Version = utils.GetPodContainerImageVersion(adv.Name, &job.Spec.Template.Spec)
		podSetStatus.Available = job.Status.Succeeded
		podSetStatus.Desired = *job.Spec.Completions
		podSetStatus.Update = &job.Status.Succeeded
		podSetStatus.Current = &job.Status.Succeeded
		podSetStatus.Ready = &job.Status.Succeeded

		status.PodSets = append(status.PodSets, podSetStatus)

		// count status
		status.Available += podSetStatus.Available
		status.Desired += podSetStatus.Desired

		// add update replicas
		updatedReplicas += job.Status.Succeeded

		if job.Status.Succeeded < *job.Spec.Completions {
			generationEqual = false
		}
	}
	return unusedObjects, status, updatedReplicas, generationEqual
}

func (w *worker) dealAggreStatus(ctx context.Context, status *workloadv1beta1.AdvDeploymentAggrStatus, generationEquanl bool, updatedReplicas int32, unUseObj []rtclient.Object) {
	sort.Slice(status.PodSets, func(i, j int) bool {
		return status.PodSets[i].Name < status.PodSets[j].Name
	})

	status.Version = removeDuplicatedVersion(status.PodSets)
	owners, ok := symctx.GetValue(ctx, types.ContextKeyAdvdeploymentOwnerRes).([]string)
	if ok {
		status.OwnerResource = owners
	}
	// stepUpdateStatus will use this args
	symctx.WithValue(ctx, types.ContextKeyAdvdeploymentAggreStatus, status)
	symctx.WithValue(ctx, types.ContextKeyAdvdeploymentGenerationEqual, generationEquanl)

	if status.Desired == status.Available && status.UnAvailable == 0 && generationEquanl && status.Desired == updatedReplicas {
		status.Status = workloadv1beta1.AppStatusRuning
	} else {
		status.Status = workloadv1beta1.AppStatusInstalling
	}

	if status.Desired > status.Available {
		return
	}

	for _, unobj := range unUseObj {
		err := w.currentCli.Delete(unobj)
		if err != nil {
			klog.Errorf("Delete unuse %s %s/%s failed: %v", unobj.GetObjectKind().GroupVersionKind().Kind, unobj.GetNamespace(), unobj.GetName(), err)
		} else {
			klog.V(4).Infof("Delete %s %s/%s successfully", unobj.GetObjectKind().GroupVersionKind().Kind, unobj.GetNamespace(), unobj.GetName())
		}
	}
}

func (w *worker) stepUpdateStatus(ctx context.Context, req ktypes.NamespacedName, adv *workloadv1beta1.AdvDeployment) error {
	var err error
	defer func() {
		if err != nil {
			symctx.WithValue(ctx, types.ContextKeyStepStop, true)
			symctx.WithValue(ctx, types.ContextKeyRequeueAfter, requeueAfterTime)
			w.currentCli.Eventf(adv, corev1.EventTypeWarning, "Update status failed: %s", err.Error())
		}
	}()

	status, ok := symctx.GetValue(ctx, types.ContextKeyAdvdeploymentAggreStatus).(*workloadv1beta1.AdvDeploymentAggrStatus)
	if !ok {
		return errors.New("advdeployment aggrrstatus is empty, must other step is error")
	}
	generationEquanl := symctx.GetValueBool(ctx, types.ContextKeyAdvdeploymentGenerationEqual)

	obj := &workloadv1beta1.AdvDeployment{}
	err = w.currentCli.Get(req, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("can't find any advdeployment %s, don't care about it", req)
		}
		return fmt.Errorf("get adveployment %s failed: %v", req, err)
	}

	if obj.Status.ObservedGeneration == obj.ObjectMeta.Generation && equality.Semantic.DeepEqual(&obj.Status.AggrStatus, status) {
		klog.V(4).Infof("Advdeployment %s status is equal not need update", req)
		return nil
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		now := metav1.Now()
		obj.Status.LastUpdateTime = &now
		status.DeepCopyInto(&obj.Status.AggrStatus)
		// It is very useful for controller that support this field
		// without this, you might trigger a sync as a result of updating your own status.
		if generationEquanl {
			obj.Status.ObservedGeneration = obj.ObjectMeta.Generation
		} else {
			obj.Status.ObservedGeneration = obj.ObjectMeta.Generation - 1
		}

		updateErr := w.currentCli.StatusUpdate(obj)
		if updateErr == nil {
			klog.V(3).Infof("Update advdeployment %s status successfully", req)
			return nil
		}
		getErr := w.currentCli.Get(req, obj)
		if getErr != nil {
			klog.Errorf("Re-get advdeployment %s failed: %v", req, getErr)
			return getErr
		}
		if obj.Status.ObservedGeneration == obj.ObjectMeta.Generation && equality.Semantic.DeepEqual(&obj.Status.AggrStatus, status) {
			// same status not need update
			klog.V(3).Infof("Re-get advdeployment compare status is equal.")
			return nil
		}
		return updateErr
	})
	return err
}
