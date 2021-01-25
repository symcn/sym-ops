/*
Copyright 2021 symcn.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

// StatefulSetStrategy is used to communicate parameter for StatefulSetStrategyType.
type StatefulSetStrategy struct {
	Partition       *int32                `json:"partition,omitempty"`
	MaxUnavailable  *intstr.IntOrString   `json:"maxUnavailable,omitempty"`
	PodUpdatePolicy PodUpdateStrategyType `json:"podUpdatePolicy,omitempty"`
}

// AdvDeploymentUpdateStrategy advdeployment update strategy
type AdvDeploymentUpdateStrategy struct {
	// canary, blue, green
	UpgradeType         string               `json:"upgradeType,omitempty"`
	StatefulSetStrategy *StatefulSetStrategy `json:"statefulSetStrategy,omitempty"`
	MinReadySeconds     int32                `json:"minReadySeconds,omitempty"`
	Meta                map[string]string    `json:"meta,omitempty"`
	// Priorities are the rules for calculating the priority of updating pods.
	// Each pod to be updated, will pass through these terms and get a sum of weights.
	// Also, priorityStrategy can just be allowed to work with Parallel podManagementPolicy.
	// +optional
	PriorityStrategy      *UpdatePriorityStrategy `json:"priorityStrategy,omitempty"`
	Paused                bool                    `json:"paused,omitempty"`
	NeedWaitingForConfirm bool                    `json:"needWaitingForConfirm,omitempty"`
}

// AdvDeploymentSpec defines the desired state of AdvDeployment
type AdvDeploymentSpec struct {
	// Replicas is the total desired replicas of all the subsets.
	// If unspecified, defaults to 1.
	// +optional
	Replicas    *int32  `json:"replicas,omitempty"`
	ServiceName *string `json:"serviceName,omitempty"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the workload
	// will fulfill this Template, but have a unique identity from the rest
	// of the workload.
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// UpdateStrategy indicates the strategy the advDeployment use to preform the update,
	// when template is changed.
	// +optional
	UpdateStrategy AdvDeploymentUpdateStrategy `json:"updateStrategy,omitempty"`

	// Topology describes the pods distribution detail between each of subsets.
	// +optional
	Topology Topology `json:"topology,omitempty"`

	// Indicates the number of histories to be conserved.
	// If unspecified, defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
}

// Topology defines the spread detail of each subset under UnitedDeployment.
// A UnitedDeployment manages multiple homogeneous workloads which are called subset.
// Each of subsets under the UnitedDeployment is described in Topology.
type Topology struct {
	// Contains the details of each subset. Each element in this array represents one subset
	// which will be provisioned and managed by UnitedDeployment.
	// +optional
	PodSets []*PodSet `json:"podSets,omitempty"`
}

// AdvDeploymentCondition describes the state of a adv deployment at a certain point.
type AdvDeploymentCondition struct {
	// Type of deployment condition.
	Type AdvDeploymentConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// AdvDeploymentAggrStatus advdeployment aggrestatus info
type AdvDeploymentAggrStatus struct {
	OwnerResource []string            `json:"ownerResource,omitempty"`
	Status        AppStatus           `json:"status,omitempty"`
	Version       string              `json:"version,omitempty"`
	Desired       int32               `json:"desired"`
	Available     int32               `json:"available"`
	UnAvailable   int32               `json:"unAvailable"`
	PodSets       []*PodSetStatusInfo `json:"podSets,omitempty"`
}

// AdvDeploymentStatus defines the observed state of AdvDeployment
type AdvDeploymentStatus struct {
	// observedGeneration is the most recent generation observed for this workload. It corresponds to the
	// StatefulSet's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	//
	Conditions []AdvDeploymentCondition `json:"conditions,omitempty"`

	// currentRevision, if not empty, indicates the version of the workload used to generate Pods in the
	// sequence [0,currentReplicas).
	CurrentRevision string `json:"currentRevision,omitempty"`

	// updateRevision, if not empty, indicates the version of the workload used to generate Pods in the sequence
	// [replicas-updatedReplicas,replicas)
	UpdateRevision string `json:"updateRevision,omitempty"`

	//
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// collisionCount is the count of hash collisions for the workload. The workload controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	//
	AggrStatus AdvDeploymentAggrStatus `json:"aggrStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true

// AdvDeployment is the Schema for the advdeployments API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ad
// +kubebuilder:printcolumn:name="DESIRED",type="integer",JSONPath=".status.aggrStatus.desired",description="The desired number of pods."
// +kubebuilder:printcolumn:name="AVAILABLE",type="integer",JSONPath=".status.aggrStatus.available",description="The number of pods ready."
// +kubebuilder:printcolumn:name="UNAVAILABLE",type="integer",JSONPath=".status.aggrStatus.unAvailable",description="The number of pods unAvailable."
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".status.aggrStatus.version",description="The image version."
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.aggrStatus.status",description="The app run status."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. "
type AdvDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdvDeploymentSpec   `json:"spec,omitempty"`
	Status AdvDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true

// AdvDeploymentList contains a list of AdvDeployment
type AdvDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AdvDeployment `json:"items"`
}

// PodUpdateStrategyType is a string enumeration type that enumerates
// all possible ways we can update a Pod when updating application
type PodUpdateStrategyType string

// PodUpdateStrategyType enum
const (
	RecreatePodUpdateStrategyType          PodUpdateStrategyType = "ReCreate"
	InPlaceIfPossiblePodUpdateStrategyType PodUpdateStrategyType = "InPlaceIfPossible"
	InPlaceOnlyPodUpdateStrategyType       PodUpdateStrategyType = "InPlaceOnly"
)

// AdvDeploymentConditionType advdeployment condition type
type AdvDeploymentConditionType string

// These are valid conditions of a deployment.
const (
	// Available means the deployment is available, ie. at least the minimum available
	// replicas required are up and running for at least minReadySeconds.
	DeploymentAvailable AdvDeploymentConditionType = "Available"
	// Progressing means the deployment is progressing. Progress for a deployment is
	// considered when a new replica set is created or adopted, and when new pods scale
	// up or old pods scale down. Progress is not estimated for paused deployments or
	// when progressDeadlineSeconds is not specified.
	DeploymentProgressing AdvDeploymentConditionType = "Progressing"
	// ReplicaFailure is added in a deployment when one of its pods fails to be created
	// or deleted.
	DeploymentReplicaFailure AdvDeploymentConditionType = "ReplicaFailure"
)

// DeployState deployment state
type DeployState string

// Deploy state enum
const (
	Created         DeployState = "Created"
	ReconcileFailed DeployState = "ReconcileFailed"
	Reconciling     DeployState = "Reconciling"
	Available       DeployState = "Available"
	Unmanaged       DeployState = "Unmanaged"
)

// Default makes AdvDeployment an mutating webhook
// When delete, if error occurs, finalizer is a good options for us to retry and
// record the events.
func (in *AdvDeployment) Default() {
	if !in.DeletionTimestamp.IsZero() {
		return
	}

	klog.V(4).Info("AdvDeployment: ", in.GetName())
}

// ValidateCreate implements webhook.Validator
// 1. check filed regex
func (in *AdvDeployment) ValidateCreate() error {
	klog.V(4).Info("validate AdvDeployment create: ", in.GetName())

	return nil
}

// ValidateUpdate validate HelmRequest update request
// immutable fields:
// 1. ...
func (in *AdvDeployment) ValidateUpdate(old runtime.Object) error {
	klog.V(4).Info("validate HelmRequest update: ", in.GetName())

	oldHR, ok := old.(*AdvDeployment)
	if !ok {
		return fmt.Errorf("expect old object to be a %T instead of %T", oldHR, old)
	}

	return nil
}

func init() {
	SchemeBuilder.Register(&AdvDeployment{}, &AdvDeploymentList{})
}
