package appset

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/symcn/api"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errorSep = "\n"

	defaultLabelValue               = "null"
	deleteUnexpectWaitAllReadyLable = "symcn.ops.deleteUnexpectWaitAllReady"

	requeueAfterTimeError = time.Second * 10
	requeueAfterTimeGrace = time.Second * 5
)

type step func(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) error

type execFuncWithCli func(cli api.MingleClient, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) (isChanged bool, err error)

type execFuncWithClusterDefine func(deployClusterSpec *workloadv1beta1.TargetCluster, req ktypes.NamespacedName, app *workloadv1beta1.AppSet) (isChanged bool, err error)

type complexAdvdeployment struct {
	ClusterName string
	Adv         *workloadv1beta1.AdvDeployment
}

type predicateSpec struct{}

// Create returns true if the Create event should be processed
func (p *predicateSpec) Create(obj rtclient.Object) bool {
	return true
}

// Delete returns true if the Delete event should be processed
func (p *predicateSpec) Delete(obj rtclient.Object) bool {
	return true
}

// Update returns true if the Update event should be processed
func (p *predicateSpec) Update(oldObj, newObj rtclient.Object) bool {
	oldAppset := oldObj.(*workloadv1beta1.AppSet)
	newAppset := newObj.(*workloadv1beta1.AppSet)
	if equality.Semantic.DeepEqual(oldAppset.Spec, newAppset.Spec) && utils.ObjecteMetaEqual(oldAppset.GetObjectMeta(), newAppset.GetObjectMeta()) {
		return false
	}
	return true
}

// Generic returns true if the Generic event should be processed
func (p *predicateSpec) Generic(obj rtclient.Object) bool {
	return true
}

func mergeVersion(v1, v2 string) string {
	s1 := strings.Split(strings.TrimSpace(v1), types.VersionSep)
	s2 := strings.Split(strings.TrimSpace(v2), types.VersionSep)
	m := map[string]struct{}{}
	for _, v := range s1 {
		if v == "" {
			continue
		}
		m[v] = struct{}{}
	}
	for _, v := range s2 {
		if v == "" {
			continue
		}
		m[v] = struct{}{}
	}

	s := make([]int, 0, len(m))
	for k := range m {
		i, _ := strconv.Atoi(strings.TrimLeft(k, "v"))
		s = append(s, i)
	}
	sort.Ints(s)

	r := ""
	for _, k := range s {
		r = fmt.Sprintf("%s%sv%d", r, types.VersionSep, k)
	}

	return strings.Trim(r, types.VersionSep)
}

func (m *master) concurrentExecStepForEachCluster(req ktypes.NamespacedName, app *workloadv1beta1.AppSet, handler execFuncWithCli) (isChanged bool, errMsg []string) {
	var (
		changed int32
		errCh   = make(chan error, 0)
		done    = make(chan struct{}, 0)
	)

	go func() {
		for err := range errCh {
			errMsg = append(errMsg, err.Error())
		}
		close(done)
	}()

	wg := sync.WaitGroup{}
	for _, cli := range m.multiCli.GetAllConnected() {
		wg.Add(1)

		go func(cli api.MingleClient) {
			defer wg.Done()

			isChanged, err := handler(cli, req, app)
			if err != nil {
				errCh <- err
			}
			if isChanged {
				atomic.AddInt32(&changed, 1)
			}
		}(cli)
	}
	wg.Wait()
	close(errCh)
	<-done

	return changed > 0, errMsg
}

func (m *master) concurrentExecStepForAppsetSpecifyCluster(req ktypes.NamespacedName, app *workloadv1beta1.AppSet, handler execFuncWithClusterDefine) (isChanged bool, errMsg []string) {
	var (
		changed int32
		errCh   = make(chan error, 0)
		done    = make(chan struct{}, 0)
		errs    = []string{}
	)

	go func() {
		for err := range errCh {
			errs = append(errs, err.Error())
		}
		close(done)
	}()

	wg := sync.WaitGroup{}
	for _, deployClusterSpec := range app.Spec.ClusterTopology.Clusters {
		wg.Add(1)

		go func(deployClusterSpec *workloadv1beta1.TargetCluster) {
			defer wg.Done()

			isChanged, err := handler(deployClusterSpec, req, app)
			if err != nil {
				errCh <- err
				return
			}
			if isChanged {
				atomic.AddInt32(&changed, 1)
			}
		}(deployClusterSpec)
	}
	wg.Wait()
	close(errCh)
	<-done

	return changed > 0, errs
}

func (m *master) getUnexpectAdvdeploymentClusterList(req ktypes.NamespacedName, app *workloadv1beta1.AppSet) []string {
	// build expect info with app
	expectClusterInfo := map[string]struct{}{}
	for _, cluster := range app.Spec.ClusterTopology.Clusters {
		expectClusterInfo[cluster.Name] = struct{}{}
	}

	unexpectClusterList := []string{}
	for _, clusterCli := range m.multiCli.GetAll() {
		adv := &workloadv1beta1.AdvDeployment{}
		err := clusterCli.Get(req, adv)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				// ignore error, only log
				klog.Errorf("Get %s cluster Advdeployment %s failed: %v", clusterCli.GetClusterCfgInfo().GetName(), req, err)
			}
			continue
		}
		if _, ok := expectClusterInfo[clusterCli.GetClusterCfgInfo().GetName()]; !ok {
			unexpectClusterList = append(unexpectClusterList, clusterCli.GetClusterCfgInfo().GetName())
		}
	}

	return unexpectClusterList
}

func (m *master) getUnexpectAdvdeploymentClusterListSync(req ktypes.NamespacedName, app *workloadv1beta1.AppSet) []string {
	// build expect info with app
	expectZoneInfo := map[string]struct{}{}
	expectClusterInfo := map[string]struct{}{}
	for _, cluster := range app.Spec.ClusterTopology.Clusters {
		if cluster.Meta[types.LabelKeyZone] == "" {
			// !import At least one cluster unknown zone, should jump sync delete
			return nil
		}
		expectZoneInfo[cluster.Meta[types.LabelKeyZone]] = struct{}{}
		expectClusterInfo[cluster.Name] = struct{}{}
	}

	type advStatusSituation struct {
		cluster    string
		zone       string
		installing bool
	}
	actualList := []advStatusSituation{}
	for _, clusterCli := range m.multiCli.GetAll() {
		adv := &workloadv1beta1.AdvDeployment{}
		err := clusterCli.Get(req, adv)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				// ignore error, only log
				klog.Errorf("Get %s cluster Advdeployment %s failed: %v", clusterCli.GetClusterCfgInfo().GetName(), req, err)
			}
			continue
		}

		zone, ok := adv.ObjectMeta.Labels[types.LabelKeyZone]
		if !ok {
			// adv label not found zone info, no way judge zone, so skip
			continue
		}
		actualList = append(actualList, advStatusSituation{
			cluster:    clusterCli.GetClusterCfgInfo().GetName(),
			zone:       zone,
			installing: adv.ObjectMeta.Generation != adv.Status.ObservedGeneration || adv.Status.AggrStatus.Status != workloadv1beta1.AppStatusRuning,
		})
	}

	unexpectClusterList := []string{}
	for _, actual := range actualList {
		if _, ok := expectClusterInfo[actual.cluster]; ok {
			// expect cluster
			continue
		}

		if _, ok := expectZoneInfo[actual.zone]; !ok {
			// not define in Appset, can remove now
			unexpectClusterList = append(unexpectClusterList, actual.cluster)
			continue
		}

		// judge same zone other cluster all status is running
		exist := false
		installingCount := 0
		for _, a := range actualList {
			if actual.zone == a.zone && actual.cluster != a.cluster {
				exist = true
				if a.installing {
					installingCount++
				}
			}
		}

		if !exist || (exist && installingCount == 0) {
			unexpectClusterList = append(unexpectClusterList, actual.cluster)
		}
	}
	return unexpectClusterList
}

func buildAdvdeploymentWithApp(app *workloadv1beta1.AppSet, deployClusterSpec *workloadv1beta1.TargetCluster) *workloadv1beta1.AdvDeployment {
	var replica int32
	for _, v := range deployClusterSpec.PodSets {
		replica += int32(v.Replicas.IntValue())
	}

	adv := &workloadv1beta1.AdvDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        app.Name,
			Namespace:   app.Namespace,
			Labels:      makeAdvdeploymentLabel(deployClusterSpec),
			Annotations: makeAdvdeploymentAnnotation(app),
		},
	}
	if app.Spec.ServiceName != nil {
		adv.Spec.ServiceName = app.Spec.ServiceName
	}
	adv.Spec.Replicas = &replica
	app.Spec.PodSpec.DeepCopyInto(&adv.Spec.PodSpec)

	for _, set := range deployClusterSpec.PodSets {
		podSet := set.DeepCopy()
		if len(podSet.RawValues) == 0 {
			// mock rawvalues, just for test
		}
		adv.Spec.Topology.PodSets = append(adv.Spec.Topology.PodSets, podSet)
	}
	return adv
}

func makeAdvdeploymentLabel(deployClusterSpec *workloadv1beta1.TargetCluster) map[string]string {
	labels := map[string]string{}

	labels[types.ObserveMustLabelClusterName] = utils.GetMapWithDefaultValue(deployClusterSpec.Meta, types.ObserveMustLabelClusterName, deployClusterSpec.Name)
	labels[types.LabelKeyZone] = utils.GetMapWithDefaultValue(deployClusterSpec.Meta, types.LabelKeyZone, defaultLabelValue)
	return labels
}

func makeAdvdeploymentAnnotation(app *workloadv1beta1.AppSet) map[string]string {
	an := map[string]string{}

	if v, ok := app.Annotations[types.AnnotationsHpa]; ok {
		an[types.AnnotationsHpa] = v
	}
	if v, ok := app.Annotations[types.AnnotationsHpaMetrics]; ok {
		an[types.AnnotationsHpaMetrics] = v
	}

	return an
}

func isAdvdeploymentDifferent(new, old *workloadv1beta1.AdvDeployment) bool {
	if !utils.ObjecteMetaEqual(new, old) {
		return true
	}

	if !equality.Semantic.DeepEqual(new.Spec, old.Spec) {
		return true
	}
	return false
}

func checkEventLabel(name string, appName string) bool {
	// name container-api-gz01a-blue-7488db8644-8zmfh
	rep, _ := regexp.Compile(fmt.Sprintf(`^(%s)-(gz|rz)(\d+?\w)-([a-z]*?)-.*?$`, appName))
	return rep.Match([]byte(name))
}

func removeDuplicatesEvent(list []*corev1.Event) []*corev1.Event {
	v := map[string]struct{}{}
	result := make([]*corev1.Event, 0, len(list))

	for _, event := range list {
		if _, ok := v[event.Reason]; !ok {
			v[event.Reason] = struct{}{}
			result = append(result, event)
		}
	}
	return result
}

func transformWorkloadEvent(list []*corev1.Event) []*workloadv1beta1.Event {
	result := make([]*workloadv1beta1.Event, 0, len(list))
	for _, event := range list {
		result = append(result, &workloadv1beta1.Event{
			Message:         event.Message,
			Reason:          event.Reason,
			Type:            event.Type,
			FirstSeen:       event.FirstTimestamp,
			LastSeen:        event.LastTimestamp,
			Count:           event.Count,
			SourceComponent: event.Source.Component,
			Name:            event.InvolvedObject.Name,
		})
	}
	return result
}
