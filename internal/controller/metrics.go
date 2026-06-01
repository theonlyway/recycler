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

var (
	// recycleTotal counts every pod deleted by the recycler controller.
	// Labels: namespace, recycler (CR name), pod.
	recycleTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recycler_pod_recycles_total",
			Help: "Total number of pods recycled by the recycler controller.",
		},
		[]string{"namespace", "recycler", "pod"},
	)

	// cpuBreachDuration observes how long a pod spent above the CPU threshold
	// before it was deleted (i.e. the elapsed time since the breach annotation).
	// Labels: namespace, recycler (CR name), pod.
	cpuBreachDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "recycler_cpu_breach_duration_seconds",
			Help:    "Duration in seconds between CPU threshold breach annotation and pod deletion.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "recycler", "pod"},
	)

	// cpuBreachesTotal counts the number of times a pod has had its CPU
	// breach annotation set (i.e. crossed the threshold).
	// Labels: namespace, recycler (CR name), pod.
	cpuBreachesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recycler_cpu_threshold_breaches_total",
			Help: "Total number of CPU threshold breach events detected across monitored pods.",
		},
		[]string{"namespace", "recycler", "pod"},
	)

	// podCPUUtilization tracks the current rolling-average CPU utilisation for
	// each monitored pod, as computed from the metrics history window.
	// Labels: namespace, pod.
	podCPUUtilization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "recycler_pod_cpu_utilization_percent",
			Help: "Current rolling-average CPU utilisation percentage for each monitored pod.",
		},
		[]string{"namespace", "pod"},
	)
)

func init() {
	// metrics.Registry is controller-runtime's global Prometheus registry,
	// which is already wired to the /metrics endpoint by the manager.
	metrics.Registry.MustRegister(
		recycleTotal,
		cpuBreachDuration,
		cpuBreachesTotal,
		podCPUUtilization,
	)
}
