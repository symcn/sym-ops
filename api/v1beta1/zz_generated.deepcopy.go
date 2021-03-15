// +build !ignore_autogenerated

/*
copyright 2021 symcn.

licensed under the apache license, version 2.0 (the "license");
you may not use this file except in compliance with the license.
you may obtain a copy of the license at

    http://www.apache.org/licenses/license-2.0

unless required by applicable law or agreed to in writing, software
distributed under the license is distributed on an "as is" basis,
without warranties or conditions of any kind, either express or implied.
see the license for the specific language governing permissions and
limitations under the license.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeployment) DeepCopyInto(out *AdvDeployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeployment.
func (in *AdvDeployment) DeepCopy() *AdvDeployment {
	if in == nil {
		return nil
	}
	out := new(AdvDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AdvDeployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeploymentAggrStatus) DeepCopyInto(out *AdvDeploymentAggrStatus) {
	*out = *in
	if in.OwnerResource != nil {
		in, out := &in.OwnerResource, &out.OwnerResource
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PodSets != nil {
		in, out := &in.PodSets, &out.PodSets
		*out = make([]*PodSetStatusInfo, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PodSetStatusInfo)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeploymentAggrStatus.
func (in *AdvDeploymentAggrStatus) DeepCopy() *AdvDeploymentAggrStatus {
	if in == nil {
		return nil
	}
	out := new(AdvDeploymentAggrStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeploymentCondition) DeepCopyInto(out *AdvDeploymentCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeploymentCondition.
func (in *AdvDeploymentCondition) DeepCopy() *AdvDeploymentCondition {
	if in == nil {
		return nil
	}
	out := new(AdvDeploymentCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeploymentList) DeepCopyInto(out *AdvDeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AdvDeployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeploymentList.
func (in *AdvDeploymentList) DeepCopy() *AdvDeploymentList {
	if in == nil {
		return nil
	}
	out := new(AdvDeploymentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AdvDeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeploymentSpec) DeepCopyInto(out *AdvDeploymentSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.ServiceName != nil {
		in, out := &in.ServiceName, &out.ServiceName
		*out = new(string)
		**out = **in
	}
	in.PodSpec.DeepCopyInto(&out.PodSpec)
	in.UpdateStrategy.DeepCopyInto(&out.UpdateStrategy)
	in.Topology.DeepCopyInto(&out.Topology)
	if in.RevisionHistoryLimit != nil {
		in, out := &in.RevisionHistoryLimit, &out.RevisionHistoryLimit
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeploymentSpec.
func (in *AdvDeploymentSpec) DeepCopy() *AdvDeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(AdvDeploymentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeploymentStatus) DeepCopyInto(out *AdvDeploymentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]AdvDeploymentCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
	if in.CollisionCount != nil {
		in, out := &in.CollisionCount, &out.CollisionCount
		*out = new(int32)
		**out = **in
	}
	in.AggrStatus.DeepCopyInto(&out.AggrStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeploymentStatus.
func (in *AdvDeploymentStatus) DeepCopy() *AdvDeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(AdvDeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AdvDeploymentUpdateStrategy) DeepCopyInto(out *AdvDeploymentUpdateStrategy) {
	*out = *in
	if in.StatefulSetStrategy != nil {
		in, out := &in.StatefulSetStrategy, &out.StatefulSetStrategy
		*out = new(StatefulSetStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.Meta != nil {
		in, out := &in.Meta, &out.Meta
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PriorityStrategy != nil {
		in, out := &in.PriorityStrategy, &out.PriorityStrategy
		*out = new(UpdatePriorityStrategy)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AdvDeploymentUpdateStrategy.
func (in *AdvDeploymentUpdateStrategy) DeepCopy() *AdvDeploymentUpdateStrategy {
	if in == nil {
		return nil
	}
	out := new(AdvDeploymentUpdateStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AggrAppSetStatus) DeepCopyInto(out *AggrAppSetStatus) {
	*out = *in
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make([]*ClusterAppActual, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ClusterAppActual)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Pods != nil {
		in, out := &in.Pods, &out.Pods
		*out = make([]*Pod, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Pod)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.WarnEvents != nil {
		in, out := &in.WarnEvents, &out.WarnEvents
		*out = make([]*Event, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Event)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(Service)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AggrAppSetStatus.
func (in *AggrAppSetStatus) DeepCopy() *AggrAppSetStatus {
	if in == nil {
		return nil
	}
	out := new(AggrAppSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSet) DeepCopyInto(out *AppSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSet.
func (in *AppSet) DeepCopy() *AppSet {
	if in == nil {
		return nil
	}
	out := new(AppSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSetCondition) DeepCopyInto(out *AppSetCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSetCondition.
func (in *AppSetCondition) DeepCopy() *AppSetCondition {
	if in == nil {
		return nil
	}
	out := new(AppSetCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSetList) DeepCopyInto(out *AppSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AppSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSetList.
func (in *AppSetList) DeepCopy() *AppSetList {
	if in == nil {
		return nil
	}
	out := new(AppSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSetSpec) DeepCopyInto(out *AppSetSpec) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Meta != nil {
		in, out := &in.Meta, &out.Meta
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.ServiceName != nil {
		in, out := &in.ServiceName, &out.ServiceName
		*out = new(string)
		**out = **in
	}
	in.PodSpec.DeepCopyInto(&out.PodSpec)
	in.UpdateStrategy.DeepCopyInto(&out.UpdateStrategy)
	in.ClusterTopology.DeepCopyInto(&out.ClusterTopology)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSetSpec.
func (in *AppSetSpec) DeepCopy() *AppSetSpec {
	if in == nil {
		return nil
	}
	out := new(AppSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSetStatus) DeepCopyInto(out *AppSetStatus) {
	*out = *in
	if in.LastUpdateTime != nil {
		in, out := &in.LastUpdateTime, &out.LastUpdateTime
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]AppSetCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.AggrStatus.DeepCopyInto(&out.AggrStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSetStatus.
func (in *AppSetStatus) DeepCopy() *AppSetStatus {
	if in == nil {
		return nil
	}
	out := new(AppSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSetUpdateStrategy) DeepCopyInto(out *AppSetUpdateStrategy) {
	*out = *in
	if in.PriorityStrategy != nil {
		in, out := &in.PriorityStrategy, &out.PriorityStrategy
		*out = new(UpdatePriorityStrategy)
		(*in).DeepCopyInto(*out)
	}
	if in.CanaryClusters != nil {
		in, out := &in.CanaryClusters, &out.CanaryClusters
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSetUpdateStrategy.
func (in *AppSetUpdateStrategy) DeepCopy() *AppSetUpdateStrategy {
	if in == nil {
		return nil
	}
	out := new(AppSetUpdateStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartSpec) DeepCopyInto(out *ChartSpec) {
	*out = *in
	if in.RawChart != nil {
		in, out := &in.RawChart, &out.RawChart
		*out = new([]byte)
		if **in != nil {
			in, out := *in, *out
			*out = make([]byte, len(*in))
			copy(*out, *in)
		}
	}
	if in.CharURL != nil {
		in, out := &in.CharURL, &out.CharURL
		*out = new(ChartURL)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartSpec.
func (in *ChartSpec) DeepCopy() *ChartSpec {
	if in == nil {
		return nil
	}
	out := new(ChartSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartURL) DeepCopyInto(out *ChartURL) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartURL.
func (in *ChartURL) DeepCopy() *ChartURL {
	if in == nil {
		return nil
	}
	out := new(ChartURL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAppActual) DeepCopyInto(out *ClusterAppActual) {
	*out = *in
	if in.PodSets != nil {
		in, out := &in.PodSets, &out.PodSets
		*out = make([]*PodSetStatusInfo, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PodSetStatusInfo)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAppActual.
func (in *ClusterAppActual) DeepCopy() *ClusterAppActual {
	if in == nil {
		return nil
	}
	out := new(ClusterAppActual)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterTopology) DeepCopyInto(out *ClusterTopology) {
	*out = *in
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make([]*TargetCluster, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TargetCluster)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterTopology.
func (in *ClusterTopology) DeepCopy() *ClusterTopology {
	if in == nil {
		return nil
	}
	out := new(ClusterTopology)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]ServicePort, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Event) DeepCopyInto(out *Event) {
	*out = *in
	in.FirstSeen.DeepCopyInto(&out.FirstSeen)
	in.LastSeen.DeepCopyInto(&out.LastSeen)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Event.
func (in *Event) DeepCopy() *Event {
	if in == nil {
		return nil
	}
	out := new(Event)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pod) DeepCopyInto(out *Pod) {
	*out = *in
	in.StartTime.DeepCopyInto(&out.StartTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pod.
func (in *Pod) DeepCopy() *Pod {
	if in == nil {
		return nil
	}
	out := new(Pod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodSet) DeepCopyInto(out *PodSet) {
	*out = *in
	if in.NodeSelectorTerm != nil {
		in, out := &in.NodeSelectorTerm, &out.NodeSelectorTerm
		*out = new(corev1.NodeSelectorTerm)
		(*in).DeepCopyInto(*out)
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.Chart != nil {
		in, out := &in.Chart, &out.Chart
		*out = new(ChartSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Mata != nil {
		in, out := &in.Mata, &out.Mata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodSet.
func (in *PodSet) DeepCopy() *PodSet {
	if in == nil {
		return nil
	}
	out := new(PodSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodSetStatusInfo) DeepCopyInto(out *PodSetStatusInfo) {
	*out = *in
	if in.HaveDeploy != nil {
		in, out := &in.HaveDeploy, &out.HaveDeploy
		*out = new(bool)
		**out = **in
	}
	if in.Ready != nil {
		in, out := &in.Ready, &out.Ready
		*out = new(int32)
		**out = **in
	}
	if in.Update != nil {
		in, out := &in.Update, &out.Update
		*out = new(int32)
		**out = **in
	}
	if in.Current != nil {
		in, out := &in.Current, &out.Current
		*out = new(int32)
		**out = **in
	}
	if in.Running != nil {
		in, out := &in.Running, &out.Running
		*out = new(int32)
		**out = **in
	}
	if in.WarnEvent != nil {
		in, out := &in.WarnEvent, &out.WarnEvent
		*out = new(int32)
		**out = **in
	}
	if in.EndpointReady != nil {
		in, out := &in.EndpointReady, &out.EndpointReady
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodSetStatusInfo.
func (in *PodSetStatusInfo) DeepCopy() *PodSetStatusInfo {
	if in == nil {
		return nil
	}
	out := new(PodSetStatusInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodSpec) DeepCopyInto(out *PodSpec) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		*out = new(corev1.PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Chart != nil {
		in, out := &in.Chart, &out.Chart
		*out = new(ChartSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodSpec.
func (in *PodSpec) DeepCopy() *PodSpec {
	if in == nil {
		return nil
	}
	out := new(PodSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	in.InternalEndpoint.DeepCopyInto(&out.InternalEndpoint)
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Domain != nil {
		in, out := &in.Domain, &out.Domain
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServicePort) DeepCopyInto(out *ServicePort) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServicePort.
func (in *ServicePort) DeepCopy() *ServicePort {
	if in == nil {
		return nil
	}
	out := new(ServicePort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StatefulSetStrategy) DeepCopyInto(out *StatefulSetStrategy) {
	*out = *in
	if in.Partition != nil {
		in, out := &in.Partition, &out.Partition
		*out = new(int32)
		**out = **in
	}
	if in.MaxUnavailable != nil {
		in, out := &in.MaxUnavailable, &out.MaxUnavailable
		*out = new(intstr.IntOrString)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StatefulSetStrategy.
func (in *StatefulSetStrategy) DeepCopy() *StatefulSetStrategy {
	if in == nil {
		return nil
	}
	out := new(StatefulSetStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetCluster) DeepCopyInto(out *TargetCluster) {
	*out = *in
	if in.Meta != nil {
		in, out := &in.Meta, &out.Meta
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodSets != nil {
		in, out := &in.PodSets, &out.PodSets
		*out = make([]*PodSet, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PodSet)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetCluster.
func (in *TargetCluster) DeepCopy() *TargetCluster {
	if in == nil {
		return nil
	}
	out := new(TargetCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Topology) DeepCopyInto(out *Topology) {
	*out = *in
	if in.PodSets != nil {
		in, out := &in.PodSets, &out.PodSets
		*out = make([]*PodSet, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PodSet)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Topology.
func (in *Topology) DeepCopy() *Topology {
	if in == nil {
		return nil
	}
	out := new(Topology)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpdatePriorityOrderTerm) DeepCopyInto(out *UpdatePriorityOrderTerm) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpdatePriorityOrderTerm.
func (in *UpdatePriorityOrderTerm) DeepCopy() *UpdatePriorityOrderTerm {
	if in == nil {
		return nil
	}
	out := new(UpdatePriorityOrderTerm)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpdatePriorityStrategy) DeepCopyInto(out *UpdatePriorityStrategy) {
	*out = *in
	if in.OrderPriority != nil {
		in, out := &in.OrderPriority, &out.OrderPriority
		*out = make([]UpdatePriorityOrderTerm, len(*in))
		copy(*out, *in)
	}
	if in.WeightPriority != nil {
		in, out := &in.WeightPriority, &out.WeightPriority
		*out = make([]UpdatePriorityWeightTerm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpdatePriorityStrategy.
func (in *UpdatePriorityStrategy) DeepCopy() *UpdatePriorityStrategy {
	if in == nil {
		return nil
	}
	out := new(UpdatePriorityStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UpdatePriorityWeightTerm) DeepCopyInto(out *UpdatePriorityWeightTerm) {
	*out = *in
	in.MatchSelector.DeepCopyInto(&out.MatchSelector)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UpdatePriorityWeightTerm.
func (in *UpdatePriorityWeightTerm) DeepCopy() *UpdatePriorityWeightTerm {
	if in == nil {
		return nil
	}
	out := new(UpdatePriorityWeightTerm)
	in.DeepCopyInto(out)
	return out
}
