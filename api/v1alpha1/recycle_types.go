/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	//+default="Deployment"
	//+optional
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`

	// name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// apiVersion is the API version of the referent
	//+default="apps/v1"
	//+optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
}

// RecycleSpec defines the desired state of Recycle
type RecycleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ScaleTargetRef from autoscalingv2 used by the horzontal pod autoscaler for consistency
	ScaleTargetRef CrossVersionObjectReference `json:"scaleTargetRef"`
	// Average CPU utilization percent of the target resource
	AverageCpuUtilizationPercent int32 `json:"averageCpuUtilizationPercent"`
	// Duration in seconds to wait before recycling
	DurationThresholdSeconds int32 `json:"durationThresholdSeconds"`
}

// RecycleStatus defines the observed state of Recycle
type RecycleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions store the status conditions of the Recycler instances
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Recycle is the Schema for the recycles API
type Recycle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecycleSpec   `json:"spec,omitempty"`
	Status RecycleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RecycleList contains a list of Recycle
type RecycleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Recycle `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Recycle{}, &RecycleList{})
}
