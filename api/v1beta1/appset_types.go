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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// AppSet represents a union for app

// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=as
// +kubebuilder:printcolumn:name="DESIRED",type="integer",JSONPath=".status.aggrStatus.desired",description="The desired number of pods."
// +kubebuilder:printcolumn:name="AVAILABLE",type="integer",JSONPath=".status.aggrStatus.available",description="The number of pods ready."
// +kubebuilder:printcolumn:name="UNAVAILABLE",type="integer",JSONPath=".status.aggrStatus.unAvailable",description="The number of pods unAvailable."
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".status.aggrStatus.version",description="The image version."
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.aggrStatus.status",description="The app run status."
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. "
type AppSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSetSpec   `json:"spec,omitempty"`
	Status AppSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// AppStatusList implements list of AppStatus.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppSet `json:"items"`
}

// AppSetSpec contains AppSet specification
type AppSetSpec struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	Replicas    *int32            `json:"replicas,omitempty"`
	ServiceName *string           `json:"serviceName,omitempty"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the workload
	// will fulfill this Template, but have a unique identity from the rest
	// of the workload.
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// UpdateStrategy indicates the strategy the advDeployment use to preform the update,
	// when template is changed.
	// +optional
	UpdateStrategy AppSetUpdateStrategy `json:"updateStrategy,omitempty"`
	// Topology describes the pods distribution detail between each of subsets.
	// +optional
	ClusterTopology ClusterTopology `json:"clusterTopology,omitempty"`
}

type AppSetUpdateStrategy struct {
	// canary, blue, green
	UpgradeType           string                  `json:"upgradeType,omitempty"`
	MinReadySeconds       int32                   `json:"minReadySeconds,omitempty"`
	PriorityStrategy      *UpdatePriorityStrategy `json:"priorityStrategy,omitempty"`
	CanaryClusters        []string                `json:"canaryClusters,omitempty"`
	Paused                bool                    `json:"paused,omitempty"`
	NeedWaitingForConfirm bool                    `json:"needWaitingForConfirm,omitempty"`
}

type ClusterTopology struct {
	Clusters []*TargetCluster `json:"clusters,omitempty"`
}

type TargetCluster struct {
	// Target cluster name
	Name string `json:"name,omitempty"`

	// exp: zone, rack
	Meta map[string]string `json:"meta,omitempty"`

	// Contains the details of each subset. Each element in this array represents one subset
	// which will be provisioned and managed by UnitedDeployment.
	// +optional
	PodSets []*PodSet `json:"podSets,omitempty"`
}

// AppSetConditionType indicates valid conditions type of a UnitedDeployment.
type AppSetConditionType string

// UnitedDeploymentCondition describes current state of a UnitedDeployment.
type AppSetCondition struct {
	// Type of in place set condition.
	Type AppSetConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// AppSetStatus contains AppSet status
type AppSetStatus struct {
	// ObservedGeneration is the most recent generation observed for this worklod. It corresponds to the
	// worklod's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// Represents the latest available observations of a UnitedDeployment's current state.
	// +optional
	Conditions []AppSetCondition `json:"conditions,omitempty"`

	AggrStatus AggrAppSetStatus `json:"aggrStatus,omitempty"`
}

// ClusterAppActual cluster app actual info
type ClusterAppActual struct {
	Name        string              `json:"name,omitempty"`
	Desired     int32               `json:"desired,omitempty"`
	Available   int32               `json:"available,omitempty"`
	UnAvailable int32               `json:"unAvailable,omitempty"`
	PodSets     []*PodSetStatusInfo `json:"podSets,omitempty"`
}

// AggrAppSetStatus represent the app status
type AggrAppSetStatus struct {
	Status      AppStatus `json:"status,omitempty"`
	Version     string    `json:"version,omitempty"`
	Desired     int32     `json:"desired"`
	Available   int32     `json:"available"`
	UnAvailable int32     `json:"unAvailable"`

	Clusters   []*ClusterAppActual `json:"clusters,omitempty"`
	Pods       []*Pod              `json:"pods,omitempty"`
	WarnEvents []*Event            `json:"warnEvents,omitempty"`
	Service    *Service            `json:"service,omitempty"`
}

func init() {
	SchemeBuilder.Register(&AppSet{}, &AppSetList{})
}
