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

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const labelNamespace string = "recycler_namespace"
const labelPod string = "recycler_pod"

var (
	// recycleTotal counts every pod deleted by the recycler controller.
	// Labels: namespace, recycler (CR name).
	recycleTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recycler_pod_recycles_total",
			Help: "Total number of pods recycled by the recycler controller.",
		},
		[]string{labelNamespace, recyclerControllerName},
	)

	// cpuBreachDuration observes how long a pod spent above the CPU threshold
	// before it was deleted (i.e. the elapsed time since the breach annotation).
	// Labels: namespace, recycler (CR name).
	cpuBreachDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "recycler_cpu_breach_duration_seconds",
			Help: "Duration in seconds between CPU threshold breach annotation and pod deletion.",
			// Buckets cover the range of realistic recycle delays (spec.recycleDelaySeconds).
			// Default is 300s; buckets span 30s–1800s to capture most configurations.
			Buckets: []float64{30, 60, 120, 180, 300, 600, 900, 1800},
		},
		[]string{labelNamespace, recyclerControllerName},
	)

	// cpuBreachesTotal counts the number of times any pod has crossed the CPU threshold.
	// Labels: namespace, recycler (CR name).
	cpuBreachesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recycler_cpu_threshold_breaches_total",
			Help: "Total number of CPU threshold breach events detected across monitored pods.",
		},
		[]string{labelNamespace, recyclerControllerName},
	)

	// podLastRecycleTime records the Unix timestamp at which each pod was most recently
	// recycled. Query this gauge to build an audit history of which pods were terminated
	// and when. Series persist in Prometheus until the retention window expires, giving
	// a queryable log of past recycle events.
	// Labels: namespace, recycler (CR name), pod.
	podLastRecycleTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "recycler_pod_last_recycle_timestamp_seconds",
			Help: "Unix timestamp of the most recent recycle event for each pod.",
		},
		[]string{labelNamespace, recyclerControllerName, labelPod},
	)

	// podCPUUtilization tracks the current rolling-average CPU utilisation for
	// each monitored pod, as computed from the metrics history window.
	// Labels: namespace, pod.
	podCPUUtilization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "recycler_pod_cpu_utilization_percent",
			Help: "Current rolling-average CPU utilisation percentage for each monitored pod.",
		},
		[]string{labelNamespace, labelPod},
	)
)

func init() {
	// metrics.Registry is controller-runtime's global Prometheus registry,
	// which is already wired to the /metrics endpoint by the manager.
	metrics.Registry.MustRegister(
		recycleTotal,
		cpuBreachDuration,
		cpuBreachesTotal,
		podLastRecycleTime,
		podCPUUtilization,
	)
}
