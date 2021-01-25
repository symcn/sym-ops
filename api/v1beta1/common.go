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
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Pod info
type Pod struct {
	Name        string      `json:"name,omitempty"`
	Namespace   string      `json:"namespace,omitempty"`
	State       string      `json:"state,omitempty"`
	PodIP       string      `json:"podIp,omitempty"`
	NodeIP      string      `json:"nodeIp,omitempty"`
	NodeName    string      `json:"nodeName,omitempty"`
	ClusterName string      `json:"clusterName,omitempty"`
	StartTime   metav1.Time `json:"startTime,omitempty"`
}

// ServicePort is a pair of port and protocol, e.g. a service endpoint.
type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	Protocol   string `json:"protocol"`
	TargetPort int32  `json:"targetPort"`
}

// Event is a single event representation.
type Event struct {
	// A human-readable description of the status of related object.
	Message string `json:"message,omitempty"`

	// Component from which the event is generated.
	SourceComponent string `json:"sourceComponent,omitempty"`

	Name string `json:"name,omitempty"`

	// Reference to a piece of an object, which triggered an event. For example
	// "spec.containers{name}" refers to container within pod with given name, if no container
	// name is specified, for example "spec.containers[2]", then it refers to container with
	// index 2 in this pod.
	SubObject string `json:"object,omitempty"`

	// The number of times this event has occurred.
	Count int32 `json:"count,omitempty"`

	// The time at which the event was first recorded.
	FirstSeen metav1.Time `json:"firstSeen,omitempty"`

	// The time at which the most recent occurrence of this event was recorded.
	LastSeen metav1.Time `json:"lastSeen,omitempty"`

	// Short, machine understandable string that gives the reason
	// for this event being generated.
	Reason string `json:"reason,omitempty"`

	// Event type (at the moment only normal and warning are supported).
	Type string `json:"type,omitempty"`
}

// Endpoint describes an endpoint that is host and a list of available ports for that host.
type Endpoint struct {
	// Hostname, either as a domain name or IP address.
	Host string `json:"host,omitempty"`

	// List of ports opened for this endpoint on the hostname.
	Ports []ServicePort `json:"ports,omitempty"`
}

// Service service info
type Service struct {
	// InternalEndpoint of all Kubernetes services that have the same label selector as connected Replication
	// Controller. Endpoints is DNS name merged with ports.
	InternalEndpoint Endpoint `json:"internalEndpoint,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	// Label selector of the service.
	Selector map[string]string `json:"selector,omitempty"`

	// Type determines how the service will be exposed.  Valid options: ClusterIP, NodePort, LoadBalancer
	Type string `json:"type,omitempty"`

	// ClusterIP is usually assigned by the master. Valid values are None, empty string (""), or
	// a valid IP address. None can be specified for headless services when proxying is not required
	ClusterIP string `json:"clusterIP,omitempty"`

	Domain *string `json:"domain,omitempty"`
}

// ChartURL char url info
type ChartURL struct {
	URL          string `json:"url,omitempty"`
	ChartVersion string `json:"chartVersion,omitempty"`
}

// ChartSpec charspec with raw content
type ChartSpec struct {
	RawChart *[]byte   `json:"rawChart,omitempty"`
	CharURL  *ChartURL `json:"chartUrl,omitempty"`
}

// PodSpec pod spec info
type PodSpec struct {
	// support PodSet：helm, InPlaceSet，StatefulSet, deployment
	// Default value is deployment
	// +optional
	DeployType string `json:"deployType,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	Selector *metav1.LabelSelector   `json:"selector,omitempty"`
	Template *corev1.PodTemplateSpec `json:"template,omitempty"`
	Chart    *ChartSpec              `json:"chart,omitempty"`
}

// PodSetStatusInfo pod status info
type PodSetStatusInfo struct {
	Name          string `json:"name"`
	Desired       int32  `json:"desired"`
	Available     int32  `json:"available"`
	UnAvailable   int32  `json:"unAvailable,omitempty"`
	Version       string `json:"version,omitempty"`
	ClusterName   string `json:"clusterName,omitempty"`
	HaveDeploy    *bool  `json:"haveDeploy,omitempty"`
	Ready         *int32 `json:"ready,omitempty"`
	Update        *int32 `json:"update,omitempty"`
	Current       *int32 `json:"current,omitempty"`
	Running       *int32 `json:"running,omitempty"`
	WarnEvent     *int32 `json:"warnEvent,omitempty"`
	EndpointReady *int32 `json:"endpointReady,omitempty"`
}

// PodSet defines the detail of a PodSet.
type PodSet struct {
	// Indicates subset name as a DNS_LABEL, which will be used to generate
	// subset workload name prefix in the format '<deployment-name>-<subset-name>-'.
	// Name should be unique between all of the subsets under one advDeployment.
	Name string `json:"name"`

	// Indicates the node selector to form the subset. Depending on the node selector,
	// pods provisioned could be distributed across multiple groups of nodes.
	// A subset's nodeSelectorTerm is not allowed to be updated.
	// +optional
	NodeSelectorTerm *corev1.NodeSelectorTerm `json:"nodeSelectorTerm,omitempty"`

	// Indicates the number of the pod to be created under this subset. Replicas could also be
	// percentage like '10%', which means 10% of UnitedDeployment replicas of pods will be distributed
	// under this subset. If nil, the number of replicas in this subset is determined by controller.
	// Controller will try to keep all the subsets with nil replicas have average pods.
	// +optional
	Replicas *intstr.IntOrString `json:"replicas,omitempty"`

	Image string `json:"image,omitempty"`

	// the images version
	Version string `json:"version,omitempty"`

	// the override podset chart spec
	Chart *ChartSpec `json:"chart,omitempty"`

	// use for helm
	RawValues string `json:"rawValues,omitempty"`

	// exp: bule/green, rz/gz
	Mata map[string]string `json:"meta,omitempty"`
}

// AppStatus app status
type AppStatus string

// AppStatus enum
const (
	AppStatusRuning       AppStatus = "Running"
	AppStatusMigrating    AppStatus = "Migrating"
	AppStatusWorkRatioing AppStatus = "WorkRatioing"
	AppStatusScaling      AppStatus = "Scaling"
	AppStatusUpdateing    AppStatus = "Updateing"
	AppStatusInstalling   AppStatus = "Installing"
	AppStatusUnknown      AppStatus = "Unknown"
)
