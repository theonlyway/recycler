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
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +default="Deployment"
	// +kubebuilder:validation:Enum=Deployment
	// +optional
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`

	// name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// apiVersion is the API version of the referent
	// +default="apps/v1"
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
}

// PrometheusSpec configures querying an external Prometheus server to determine
// per-pod CPU utilization instead of polling the Kubernetes Metrics API and
// computing a rolling average in-process.
type PrometheusSpec struct {
	// ServerAddress is the base URL of the Prometheus server to query,
	// e.g. "http://prometheus-operated.monitoring.svc:9090".
	// +kubebuilder:validation:MinLength=1
	ServerAddress string `json:"serverAddress"`
	// Query is the PromQL query used to evaluate per-pod CPU utilization. It must return an
	// instant vector where each sample carries a "pod" label and a value representing the
	// CPU utilization percentage to compare against averageCpuUtilizationPercent.
	//
	// The query is rendered as a Go text/template with the following fields available:
	//   {{.Namespace}}      - the namespace of the target Deployment
	//   {{.Deployment}}     - the name of the target Deployment
	//   {{.PodRegex}}       - a regex alternation of the current pod names (pod1|pod2|...)
	//   {{.WindowSeconds}}  - podMetricsHistory * pollingIntervalSeconds, the averaging window
	//
	// When omitted, a default query based on cAdvisor (container_cpu_usage_seconds_total) and
	// kube-state-metrics (kube_pod_container_resource_limits) is used.
	// +optional
	Query string `json:"query,omitempty"`
	// InsecureSkipVerify disables TLS certificate verification when the ServerAddress uses HTTPS.
	// +optional
	// +default=false
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// RecyclerSpec defines the desired state of Recycler
// +kubebuilder:validation:XValidation:rule="self.metricsSource != 'prometheus' || has(self.prometheus)",message="prometheus configuration is required when metricsSource is 'prometheus'"
type RecyclerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ScaleTargetRef identifies the Deployment to monitor and recycle pods for.
	// Only Deployments are supported. The name must match an existing Deployment in the same namespace.
	ScaleTargetRef CrossVersionObjectReference `json:"scaleTargetRef"`
	// AverageCpuUtilizationPercent is the rolling-average CPU usage threshold, expressed as a percentage
	// of the pod's CPU request. When the rolling average across the configured history window exceeds
	// this value, the pod is marked for recycling.
	AverageCpuUtilizationPercent int32 `json:"averageCpuUtilizationPercent"`
	// RecycleDelaySeconds is how long to wait after a CPU breach is first detected before the pod is
	// deleted. This gives transient spikes time to recover before a recycle is triggered.
	// +optional
	// +default=300
	RecycleDelaySeconds int32 `json:"recycleDelaySeconds"`
	// PollingIntervalSeconds controls how frequently the controller polls the Kubernetes Metrics API
	// to collect CPU usage for each pod. Lower values produce a more responsive rolling average but
	// increase API server load.
	// +optional
	// +default=60
	PollingIntervalSeconds int32 `json:"pollingIntervalSeconds"`
	// PodMetricsHistory is the number of polling samples retained in the rolling window used to
	// compute the average CPU utilization. A larger window smooths out short spikes; a smaller
	// window reacts more quickly to sustained high usage.
	// +optional
	// +default=10
	PodMetricsHistory int32 `json:"podMetricsHistory"`
	// GracePeriodSeconds is the pod termination grace period passed to the Kubernetes delete call.
	// The kubelet will send SIGTERM and wait this many seconds before sending SIGKILL.
	// +optional
	// +default=30
	GracePeriodSeconds int32 `json:"gracePeriodSeconds"`
	// MetricStorageLocation controls where per-pod CPU history is stored between reconcile cycles.
	// "memory" stores data in-process (fast, zero API cost, lost on controller restart).
	// "annotation" persists history as a pod annotation (survives restarts, incurs an etcd write per poll).
	// +optional
	// +kubebuilder:validation:Enum=memory;annotation
	// +kubebuilder:default=memory
	MetricStorageLocation string `json:"metricStorageLocation"`
	// MetricsSource selects how per-pod CPU utilization is obtained.
	// "kubernetes" (default) polls the Kubernetes Metrics API on pollingIntervalSeconds and
	// computes a rolling average across podMetricsHistory samples stored per metricStorageLocation.
	// "prometheus" queries an external Prometheus server (see the prometheus field) so the averaging
	// is performed by Prometheus and no per-pod history is stored by the controller.
	// +optional
	// +kubebuilder:validation:Enum=kubernetes;prometheus
	// +kubebuilder:default=kubernetes
	MetricsSource string `json:"metricsSource"`
	// Prometheus configures the external Prometheus server to query when metricsSource is "prometheus".
	// It is required when metricsSource is "prometheus" and ignored otherwise.
	// +optional
	Prometheus *PrometheusSpec `json:"prometheus,omitempty"`
	// MetricsRetentionSeconds controls how long per-pod gauge series are retained in the /metrics
	// endpoint after a pod is terminated. Series linger for this duration to allow at least one
	// Prometheus scrape to capture the final value before they are removed. Set to 0 to delete
	// series immediately upon pod termination.
	// +optional
	// +default=300
	MetricsRetentionSeconds int32 `json:"metricsRetentionSeconds"`
}

// RecyclerStatus defines the observed state of Recycler
type RecyclerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Recycler is the Schema for the recyclers API
type Recycler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecyclerSpec   `json:"spec,omitempty"`
	Status RecyclerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RecyclerList contains a list of Recycler
type RecyclerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Recycler `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(GroupVersion, &Recycler{}, &RecyclerList{})
		return nil
	})
}
