package advdeployment

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/symcn/api"
	workloadv1beta1 "github.com/symcn/sym-ops/api/v1beta1"
	"github.com/symcn/sym-ops/pkg/resource"
	"github.com/symcn/sym-ops/pkg/types"
	"github.com/symcn/sym-ops/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	rtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	requeueAfterTime = 5 * time.Second
	convertFactory   = map[string]convert{}
)

type step func(ctx context.Context, req ktypes.NamespacedName, app *workloadv1beta1.AdvDeployment) error
type convert func(obj *unstructured.Unstructured, isHpaEnable bool) (rtclient.Object, resource.Option, int32, error)

type predicateSpec struct{}

func (w *worker) initialization() {
	convertFactory = map[string]convert{
		types.ServiceKind:     w.convertToSvc,
		types.DeploymentKind:  w.convertToDeployment,
		types.StatefulSetKind: w.convertToStatefulSet,
		types.JobKind:         w.convertToJob,
	}
}

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
	oldAdv := oldObj.(*workloadv1beta1.AdvDeployment)
	newAdv := newObj.(*workloadv1beta1.AdvDeployment)
	if equality.Semantic.DeepEqual(oldAdv.Spec, newAdv.Spec) && utils.ObjecteMetaEqual(oldAdv.GetObjectMeta(), newAdv.GetObjectMeta()) {
		return false
	}
	return true
}

// Generic returns true if the Generic event should be processed
func (p *predicateSpec) Generic(obj rtclient.Object) bool {
	return true
}

type transformNamespace struct{}

func (t *transformNamespace) Create(obj rtclient.Object, queue api.WorkQueue) {
	requeue(obj, queue)
}

func (t *transformNamespace) Update(oldObj, newObj rtclient.Object, queue api.WorkQueue) {
	requeue(newObj, queue)
}

func (t *transformNamespace) Delete(obj rtclient.Object, queue api.WorkQueue) {
	requeue(obj, queue)
}

func (t *transformNamespace) Generic(obj rtclient.Object, queue api.WorkQueue) {
	requeue(obj, queue)
}

func requeue(obj rtclient.Object, queue api.WorkQueue) {
	labels := obj.GetLabels()
	if _, ok := labels[types.ObserveMustLabelClusterName]; !ok {
		return
	}
	name := labels[types.ObserveMustLabelAppName]
	if name == "" {
		return
	}

	queue.Add(ktypes.NamespacedName{
		Name:      name,
		Namespace: obj.GetNamespace(),
	})
}

func getCharInfo(podSet *workloadv1beta1.PodSet, adv *workloadv1beta1.AdvDeployment) (chartURL, chartVersion string, rawChart []byte) {
	if podSet.Chart != nil {
		if podSet.Chart.RawChart != nil {
			rawChart = *podSet.Chart.RawChart
		}
		if podSet.Chart.CharURL != nil {
			chartURL = podSet.Chart.CharURL.URL
			chartVersion = podSet.Chart.CharURL.ChartVersion
		}

		// if podset char info is empty, use global char into
		if rawChart != nil || chartURL != "" || chartVersion != "" {
			return
		}
	}

	if adv.Spec.PodSpec.Chart.RawChart != nil {
		rawChart = *adv.Spec.PodSpec.Chart.RawChart
	}
	if adv.Spec.PodSpec.Chart.CharURL != nil {
		chartURL = adv.Spec.PodSpec.Chart.CharURL.URL
		chartVersion = adv.Spec.PodSpec.Chart.CharURL.ChartVersion
	}
	return
}

type hpaSpec struct {
	Enable      bool  `json:"enable,omitempty"`
	MaxReplicas int32 `json:"max_replicas,omitempty"`
	MinReplicas int32 `json:"min_replicas,omitempty"`
}

func getHpaSpecEnable(m map[string]string) bool {
	hpaAnnotation, ok := m[types.AnnotationsHpa]
	if !ok {
		return false
	}

	hs := &hpaSpec{}
	err := json.Unmarshal([]byte(hpaAnnotation), hs)
	if err != nil {
		klog.Errorf("Unmarshal hpaAnnotation %s failed: %v", hpaAnnotation, err)
		return false
	}
	return hs.Enable
}

func getFormattedName(kind string, obj rtclient.Object) string {
	return fmt.Sprintf("%s:%s/%s", kind, obj.GetNamespace(), obj.GetName())
}

func (w *worker) convertToSvc(obj *unstructured.Unstructured, isHpaEnable bool) (rtclient.Object, resource.Option, int32, error) {
	svc := &corev1.Service{}
	err := w.currentCli.GetCtrlRtManager().GetScheme().Convert(obj, svc, nil)
	if err != nil {
		return nil, resource.Option{}, 0, fmt.Errorf("Convert to Service failed: %v", err)
	}
	return svc, resource.Option{IsRecreate: w.conf.Debug}, 0, nil
}

func (w *worker) convertToDeployment(obj *unstructured.Unstructured, isHpaEnable bool) (rtclient.Object, resource.Option, int32, error) {
	deploy := &appsv1.Deployment{}
	err := w.currentCli.GetCtrlRtManager().GetScheme().Convert(obj, deploy, nil)
	if err != nil {
		return nil, resource.Option{}, 0, fmt.Errorf("Convert to Deployment failed: %v", err)
	}

	if deploy.Spec.RevisionHistoryLimit == nil {
		if w.conf.RevisionHistoryLimit > 0 {
			deploy.Spec.RevisionHistoryLimit = &w.conf.RevisionHistoryLimit
		} else {
			deploy.Spec.RevisionHistoryLimit = &defaultRevisionHistoryLimit
		}
	}
	if deploy.Spec.ProgressDeadlineSeconds == nil {
		if w.conf.ProgressDeadlineSeconds > 0 {
			deploy.Spec.ProgressDeadlineSeconds = &w.conf.ProgressDeadlineSeconds
		} else {
			deploy.Spec.ProgressDeadlineSeconds = &defaultProgressDeadlineSeconds
		}
	}
	if deploy.Spec.Selector == nil {
		deploy.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: deploy.Spec.Template.Labels,
		}
	}

	return deploy, resource.Option{IsRecreate: w.conf.Debug, IsIgnoreReplicas: isHpaEnable}, utils.TransInt32Ptr2Int32(deploy.Spec.Replicas, 1), nil
}

func (w *worker) convertToStatefulSet(obj *unstructured.Unstructured, isHpaEnable bool) (rtclient.Object, resource.Option, int32, error) {
	statefulset := &appsv1.StatefulSet{}
	err := w.currentCli.GetCtrlRtManager().GetScheme().Convert(obj, statefulset, nil)
	if err != nil {
		return nil, resource.Option{}, 0, fmt.Errorf("Convert to StatefulSet failed: %v", err)
	}

	if statefulset.Spec.RevisionHistoryLimit == nil {
		if w.conf.RevisionHistoryLimit > 0 {
			statefulset.Spec.RevisionHistoryLimit = &w.conf.RevisionHistoryLimit
		} else {
			statefulset.Spec.RevisionHistoryLimit = &defaultRevisionHistoryLimit
		}
	}

	return statefulset, resource.Option{IsRecreate: w.conf.Debug, IsIgnoreReplicas: isHpaEnable}, utils.TransInt32Ptr2Int32(statefulset.Spec.Replicas, 1), nil
}

func (w *worker) convertToJob(obj *unstructured.Unstructured, isHpaEnable bool) (rtclient.Object, resource.Option, int32, error) {
	job := &batchv1.Job{}
	err := w.currentCli.GetCtrlRtManager().GetScheme().Convert(obj, job, nil)
	if err != nil {
		return nil, resource.Option{}, 0, fmt.Errorf("Convert to Job failed: %v", err)
	}
	return job, resource.Option{IsRecreate: w.conf.Debug, IsIgnoreReplicas: false}, 0, nil
}

func parseMetrics(annotations map[string]string, objectName string) []v2beta2.MetricSpec {
	metricsSlice := getHpaMetrics(annotations)
	if len(metricsSlice) == 0 {
		klog.V(5).Infof("Annotation don't have metrics")
		return nil
	}

	metrics := make([]v2beta2.MetricSpec, 0, 2)
	for _, m := range metricsSlice {
		var metric *v2beta2.MetricSpec
		switch m.ResourceName {
		case string(v1.ResourceCPU):
			metric = createResourceMetric(v1.ResourceCPU, m.MetricType, m.MetricValue, objectName)
		case string(v1.ResourceMemory):
			metric = createResourceMetric(v1.ResourceMemory, m.MetricType, m.MetricValue, objectName)
		default:
		}

		if metric != nil {
			metrics = append(metrics, *metric)
		}
	}

	return metrics
}

type hpaMetric struct {
	ResourceName string `json:"resource,omitempty"`
	MetricType   string `json:"metric_type,omitempty"`
	MetricValue  string `json:"metric_value,omitempty"`
}

func getHpaMetrics(m map[string]string) []*hpaMetric {
	org := m[types.AnnotationsHpaMetrics]
	if org == "" {
		return nil
	}

	metrics := []*hpaMetric{}
	err := json.Unmarshal([]byte(org), &metrics)
	if err != nil {
		klog.Errorf("unmarshal hpaMetric %s failed: %v", org, err)
		return nil
	}
	return metrics
}

func createResourceMetric(resourceName v1.ResourceName, metricType string, metricValue string, deployName string) *v2beta2.MetricSpec {
	if metricType == "" || metricValue == "" {
		klog.Errorf("Invalid resource metricType and metricValue is empty")
		return nil
	}

	switch metricType {
	case types.MetricAverageUtilization:
		val, err := strconv.ParseInt(metricValue, 10, 32)
		if err != nil {
			klog.Errorf("Invalid resource metricValue %s for deploy %s", metricValue, deployName)
			return nil
		}
		targetValue := int32(val)
		if targetValue <= 0 || targetValue > 100 {
			klog.Errorf("Invalid resource metric value %d for deploy %s should be a percentage value between [1, 100]", targetValue, deployName)
			return nil
		}
		return &v2beta2.MetricSpec{
			Type: v2beta2.ResourceMetricSourceType,
			Resource: &v2beta2.ResourceMetricSource{
				Name: resourceName,
				Target: v2beta2.MetricTarget{
					Type:               v2beta2.UtilizationMetricType,
					AverageUtilization: &targetValue,
				},
			},
		}
	case types.MetricAverageValue:
		targetValue, err := kresource.ParseQuantity(metricValue)
		if err != nil {
			klog.Errorf("Invalid resource metric value %s for deploy %s: %v", metricValue, deployName, err)
			return nil
		}
		return &v2beta2.MetricSpec{
			Type: v2beta2.ResourceMetricSourceType,
			Resource: &v2beta2.ResourceMetricSource{
				Name: resourceName,
				Target: v2beta2.MetricTarget{
					Type:         v2beta2.AverageValueMetricType,
					AverageValue: &targetValue,
				},
			},
		}

	default:
		klog.Warningf("Invalid resource metric type %s for deploy %s", metricType, deployName)
		return nil
	}
}

func equalHpa(hpa1, hpa2 *v2beta2.HorizontalPodAutoscaler) bool {
	if !utils.ObjecteMetaEqual(hpa1, hpa2) {
		return false
	}
	if !equality.Semantic.DeepEqual(hpa1.Spec, hpa2.Spec) {
		return false
	}
	return true
}

func removeDuplicatedVersion(list []*workloadv1beta1.PodSetStatusInfo) string {
	found := map[string]struct{}{}
	versions := []string{}
	for _, item := range list {
		if item.Version != "" {
			if _, ok := found[item.Version]; !ok {
				found[item.Version] = struct{}{}
				versions = append(versions, item.Version)
			}
		}
	}

	sort.Strings(versions)
	return strings.Join(versions, types.VersionSep)
}

func isUnunseObject(kind string, obj rtclient.Object, owners []string) bool {
	if len(owners) == 0 {
		// if owners is empty, shouldn't mark unused
		return false
	}

	name := getFormattedName(kind, obj)
	for _, item := range owners {
		if name == item {
			return false
		}
	}
	return true
}
