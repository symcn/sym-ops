package resource

import (
	"context"
	"fmt"
	"reflect"

	"github.com/symcn/api"
	"github.com/symcn/sym-ops/pkg/resource/patch"
	"github.com/symcn/sym-ops/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// DesiredState DesiredState type
type DesiredState string

// DesiredState enum defaine
const (
	DesiredStatePresent DesiredState = "present"
	DesiredStateAbsent  DesiredState = "absent"
)

type Option struct {
	DesiredState     DesiredState
	IsRecreate       bool
	IsIgnoreReplicas bool
}

func Reconcile(ctx context.Context, cli api.MingleClient, desired rtclient.Object, opt Option) (bool, error) {
	if opt.DesiredState == "" {
		opt.DesiredState = DesiredStatePresent
	}

	current := desired.DeepCopyObject().(rtclient.Object)
	key := rtclient.ObjectKeyFromObject(desired)
	err := cli.Get(key, current)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("Getting resource %s failed: %v", key, err)
		}

		if opt.DesiredState == DesiredStatePresent {
			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				klog.Errorf("Failed to set last applied annotation %s: %v", key, err)
			}
			if err := cli.Create(desired); err != nil {
				return false, fmt.Errorf("Create resource %s failed: %v", key, err)
			}
			klog.V(4).Infof("Create resource %s successful.", key)
		}
		return true, nil
	}

	if opt.DesiredState == DesiredStateAbsent {
		if err := cli.Delete(current); err != nil {
			return false, fmt.Errorf("Deleting resource %s failed: %v", key, err)
		}
		klog.V(4).Infof("Delete resource %s successful.", key)
		return true, nil
	}

	if opt.DesiredState != DesiredStatePresent {
		return false, nil
	}

	calcOpts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}
	if _, ok := desired.(*appsv1.StatefulSet); ok {
		calcOpts = append(calcOpts, patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus())
	}

	copts, stop, err := patchHandle(current, desired, key, opt)
	if err != nil || !stop {
		return false, err
	}
	if len(copts) > 0 {
		calcOpts = append(calcOpts, copts...)
	}

	// Need to set this before resourceversion is set, as it would constantly change otherwise
	if err = patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
		klog.Errorf("Set last applied annotation %s failed: %v", key, err)
	}
	if _, isJob := desired.(*batchv1.Job); isJob {
		return true, updateOrDeleteAndCreateJob(cli, current, desired, key)
	}

	if opt.IsRecreate {
		return true, recreate(cli, current, desired, key)
	}
	return true, retryUpdate(cli, current, desired, key)
}

func patchHandle(current, desired rtclient.Object, key ktypes.NamespacedName, opt Option) (calcOpts []patch.CalculateOption, stop bool, err error) {
	if !labelEqual(desired, current) {
		return
	}

	calcOpts = []patch.CalculateOption{}
	if opt.IsIgnoreReplicas {
		switch desired.(type) {
		case *appsv1.Deployment:
			desiredDeploy := desired.(*appsv1.Deployment)
			currentDeploy := current.(*appsv1.Deployment)
			if utils.TransInt32Ptr2Int32(desiredDeploy.Spec.Replicas, 1) > utils.TransInt32Ptr2Int32(currentDeploy.Spec.Replicas, 1) {
				calcOpts = append(calcOpts, patch.IgnoreDeployReplicasFields())
			}
		case *appsv1.StatefulSet:
			desiredSts := desired.(*appsv1.StatefulSet)
			currentSts := current.(*appsv1.StatefulSet)
			if utils.TransInt32Ptr2Int32(desiredSts.Spec.Replicas, 1) > utils.TransInt32Ptr2Int32(currentSts.Spec.Replicas, 1) {
				calcOpts = append(calcOpts, patch.IgnoreStatefulSetReplicasFields())
			}
		}
	}

	patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, calcOpts...)
	if err != nil {
		return nil, false, fmt.Errorf("Couldn't not match object %s err: %v", key, err)
	}
	if patchResult.IsEmpty() {
		klog.V(4).Infof("resource %s unchanged is in sync", key)
		return nil, true, nil
	}
	klog.V(5).Infof("resource %s diffs patch %s", key, string(patchResult.Patch))
	return calcOpts, false, nil
}

func updateOrDeleteAndCreateJob(cli api.MingleClient, current, desired rtclient.Object, key ktypes.NamespacedName) error {
	err := cli.Update(desired)
	if err == nil {
		klog.V(4).Infof("Update resource %s successful.", key)
		return nil
	}
	if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
		klog.V(5).Infof("resource %s needs to be re-created: %v", key, err)
		//https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#controlling-how-the-garbage-collector-deletes-dependents
		propagationPolicy := metav1.DeletePropagationBackground
		err = cli.Delete(current, &rtclient.DeleteOptions{PropagationPolicy: &propagationPolicy})
		if err != nil {
			return fmt.Errorf("couldn't delete resource %s", key)
		}
		klog.V(5).Infof("Delete resource %s successful", key)
		err = cli.Create(desired)
		if err != nil {
			return fmt.Errorf("Re-create resource %s failed: %v", key, err)
		}
		klog.V(4).Infof("Re-create resource %s successful", key)
		return nil
	}
	return fmt.Errorf("Update resource %s failed: %v", key, err)
}

func recreate(cli api.MingleClient, current, desired rtclient.Object, key ktypes.NamespacedName) error {
	metaAccessor := meta.NewAccessor()
	currentResourceVersion, err := metaAccessor.ResourceVersion(current)
	if err != nil {
		return fmt.Errorf("get resource %s resourceversion failed: %v", key, err)
	}

	err = metaAccessor.SetResourceVersion(desired, currentResourceVersion)
	if err != nil {
		return fmt.Errorf("set resource %s resourceversion %s failed:%v", key, currentResourceVersion, err)
	}
	prepareResourceForUpdate(current, desired)
	return updateOrDeleteAndCreateResource(cli, current, desired, key)
}

func retryUpdate(cli api.MingleClient, current, desired rtclient.Object, key ktypes.NamespacedName) error {
	metaAccessor := meta.NewAccessor()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentResourceVersion, err := metaAccessor.ResourceVersion(current)
		if err != nil {
			return fmt.Errorf("get resource %s resourceversion failed: %v", key, err)
		}

		err = metaAccessor.SetResourceVersion(desired, currentResourceVersion)
		if err != nil {
			return fmt.Errorf("set resource %s resourceversion %s failed:%v", key, currentResourceVersion, err)
		}
		prepareResourceForUpdate(current, desired)

		updateErr := cli.Update(desired)
		if updateErr != nil {
			return fmt.Errorf("update resource %s failed: %v", key, updateErr)
		}

		getErr := cli.Get(key, current)
		if getErr != nil {
			return fmt.Errorf("update get resource %s failed: %v", key, err)
		}
		return updateErr
	})
}

func prepareResourceForUpdate(current, desired rtclient.Object) {
	switch desired.(type) {
	case *corev1.Service:
		svc := desired.(*corev1.Service)
		svc.Spec.ClusterIP = current.(*corev1.Service).Spec.ClusterIP
	}
}

func updateOrDeleteAndCreateResource(cli api.MingleClient, current, desired rtclient.Object, key ktypes.NamespacedName) error {
	err := cli.Update(desired)
	if err == nil {
		klog.V(4).Infof("Update resource %s successful.", key)
		return nil
	}
	if apierrors.IsConflict(err) || apierrors.IsInvalid(err) {
		klog.V(5).Infof("resource %s needs to be re-created: %v", key, err)
		err = cli.Delete(current)
		if err != nil {
			return fmt.Errorf("couldn't delete resource %s", key)
		}
		klog.V(5).Infof("Delete resource %s successful", key)
		err = cli.Create(desired)
		if err != nil {
			return fmt.Errorf("Re-create resource %s failed: %v", key, err)
		}
		klog.V(4).Infof("Re-create resource %s successful", key)
		return nil
	}
	return fmt.Errorf("Update resource %s failed: %v", key, err)
}

func labelEqual(obj1, obj2 metav1.Object) bool {
	if !reflect.DeepEqual(obj1.GetLabels(), obj2.GetLabels()) {
		klog.V(4).Infof("object %s/%s labels different.", obj1.GetNamespace(), obj1.GetName())
		return false
	}
	return true
}
