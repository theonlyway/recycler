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
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
	"k8s.io/client-go/util/retry"
)

const monitorControllerName = "monitor"

const (
	StorageMemory     string = "memory"
	StorageAnnotation string = "annotation"
)

// Exported thread-safe in-memory storage for metrics
var InMemoryMetricsStorage sync.Map

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
	Config   *rest.Config
}

// PodCPUUsage represents the CPU usage of a pod
type PodCPUUsage struct {
	PodName       string
	CPUUsage      resource.Quantity // Raw CPU usage
	CPULimit      resource.Quantity // CPU limit
	CPUPercentage float64           // Percentage CPU utilization
	Timestamp     time.Time         // Timestamp of the metrics
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;update;patch;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Monitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("monitor", req.NamespacedName)

	// Fetch the Recycler instance
	recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
	err := r.Get(ctx, req.NamespacedName, recycler)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Recycler resource not found. Ignoring since object must be deleted", "controller", monitorControllerName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Recycle", "controller", monitorControllerName)
		return ctrl.Result{}, err
	}

	// Fetch the resource type using ScaleTargetRef
	switch kind := recycler.Spec.ScaleTargetRef.Kind; kind {
	case "Deployment":
		// Fetch the target deployment using ScaleTargetRef
		deployment := &appsv1.Deployment{}
		log.V(1).Info("Retrieving pods in target deployment", "controller", monitorControllerName, "deployment", recycler.Spec.ScaleTargetRef.Name)
		deploymentKey := client.ObjectKey{
			Namespace: recycler.Namespace,
			Name:      recycler.Spec.ScaleTargetRef.Name,
		}
		if err := r.Get(ctx, deploymentKey, deployment); err != nil {
			log.Error(err, "Failed to fetch target deployment", "controller", monitorControllerName, "deployment", deploymentKey)
			return ctrl.Result{}, err
		}
		// Fetch the pods in the deployment
		podList := &corev1.PodList{}
		listOptions := []client.ListOption{
			client.InNamespace(deployment.Namespace),
			client.MatchingLabels(deployment.Spec.Selector.MatchLabels),
		}
		if err := r.List(ctx, podList, listOptions...); err != nil {
			log.Error(err, "Failed to list pods in target deployment", "controller", monitorControllerName, "deployment", deploymentKey)
			return ctrl.Result{}, err
		}
		// Fetch the metrics for the pods in the deployment
		metricsClient := resourceclient.NewForConfigOrDie(r.Config).PodMetricses(deployment.Namespace)
		podMetricsList, err := fetchPodMetrics(ctx, metricsClient, deployment.Namespace, deployment.Spec.Selector.MatchLabels, deployment.Spec.Template, log)
		if err != nil {
			log.Error(err, "Failed to fetch metrics for pods in target deployment", "controller", monitorControllerName, "deployment", deploymentKey)
			return ctrl.Result{}, err
		}

		for _, podCPU := range podMetricsList {
			// Update the pod's metrics history based on storage location
			if err := updatePodMetricsHistory(ctx, r, podCPU.PodName, deployment.Namespace, podCPU, recycler.Spec.PodMetricsHistory, log, recycler.Spec.MetricStorageLocation); err != nil {
				log.Error(err, "Failed to update pod metrics history", "podName", podCPU.PodName)
			}
		}

		for _, pod := range podList.Items {
			// Fetch the metrics history based on storage location
			metricsHistory, err := fetchPodMetricsHistory(ctx, r, &pod, log, recycler.Spec.MetricStorageLocation)
			if err != nil {
				log.Error(err, "Failed to fetch metrics history", "podName", pod.Name)
				continue
			}

			// Check if there are enough data points
			if len(metricsHistory) < int(recycler.Spec.PodMetricsHistory) {
				log.V(1).Info("Not enough data points for pod, skipping", "podName", pod.Name)
				continue
			}

			// Check threshold and annotate if breached
			if err := checkPodMetricsAnnotation(ctx, r, recycler, &pod, metricsHistory, recycler.Spec.AverageCpuUtilizationPercent, log); err != nil {
				log.Error(err, "Failed to check threshold and annotate pod", "podName", pod.Name)
			}
		}
	default:
		log.Info("Unsupported resource type", "controller", monitorControllerName, "kind", kind)
	}

	// Log all in-memory metrics at the end of the reconciliation loop
	logInMemoryMetrics(r.Log)

	return ctrl.Result{RequeueAfter: time.Duration(recycler.Spec.PollingIntervalSeconds) * time.Second}, nil
}

func logInMemoryMetrics(log logr.Logger) {
	InMemoryMetricsStorage.Range(func(key, value interface{}) bool {
		metricsHistory := value.([]PodCPUUsage)
		for i, metric := range metricsHistory {
			log.V(1).Info("In-memory metric stored", "key", key, "index", i, "podName", metric.PodName, "CPUUsage", metric.CPUUsage.String(), "CPULimit", metric.CPULimit.String(), "CPUPercentage", metric.CPUPercentage, "Timestamp", metric.Timestamp)
		}
		return true
	})
}

func fetchPodMetrics(ctx context.Context, metricsClient resourceclient.PodMetricsInterface, namespace string, labelSelector map[string]string, podTemplate corev1.PodTemplateSpec, log logr.Logger) ([]PodCPUUsage, error) {
	// Create a label selector from the provided labels
	selector := labels.SelectorFromSet(labelSelector).String()

	// Fetch the pod metrics using the Kubernetes Metrics API client
	podMetricsList, err := metricsClient.List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		log.Error(err, "Failed to fetch pod metrics", "namespace", namespace, "labelSelector", labelSelector)
		return nil, err
	}

	if len(podMetricsList.Items) == 0 {
		log.Info("No pod metrics returned", "namespace", namespace, "labelSelector", labelSelector)
		return nil, fmt.Errorf("no pod metrics returned from resource metrics API")
	}

	// Process the metrics and calculate CPU utilization for each pod
	podCPUUsages := make([]PodCPUUsage, 0, len(podMetricsList.Items))
	for _, podMetrics := range podMetricsList.Items {
		// Sum the CPU usage across all containers in the pod
		totalCPUUsage := resource.Quantity{}
		for _, container := range podMetrics.Containers {
			if cpuUsage, found := container.Usage[corev1.ResourceCPU]; found {
				totalCPUUsage.Add(cpuUsage)
			} else {
				log.Info("Missing CPU usage metric for container", "containerName", container.Name, "podName", podMetrics.Name)
			}
		}

		// Get the CPU limit from the pod template
		totalCPULimit := resource.Quantity{}
		for _, container := range podTemplate.Spec.Containers {
			if container.Resources.Limits != nil {
				if cpuLimit, found := container.Resources.Limits[corev1.ResourceCPU]; found {
					totalCPULimit.Add(cpuLimit)
				}
			}
		}

		// Calculate the percentage CPU utilization
		var cpuUtilization float64
		if totalCPULimit.MilliValue() > 0 {
			// Convert millicores to cores by dividing by 1000
			cpuUtilization = (float64(totalCPUUsage.MilliValue()) / float64(totalCPULimit.MilliValue())) * 100
		} else {
			log.Info("Pod CPU limit is 0, skipping CPU utilization calculation", "podName", podMetrics.Name)
			cpuUtilization = 0 // No CPU limit defined
		}

		// Format the CPU utilization to two decimal places
		cpuUtilization = math.Round(cpuUtilization*100) / 100

		// Append the pod's CPU utilization to the result list, including the current timestamp
		podCPUUsages = append(podCPUUsages, PodCPUUsage{
			PodName:       podMetrics.Name,
			CPUUsage:      totalCPUUsage,
			CPULimit:      totalCPULimit,
			CPUPercentage: cpuUtilization,
			Timestamp:     podMetrics.Timestamp.Time, // Use the metrics query timestamp
		})
	}

	return podCPUUsages, nil
}

// UpdatePodMetricsHistory updates the pod's metrics history in memory or annotations
func updatePodMetricsHistory(ctx context.Context, r *MonitorReconciler, podName string, namespace string, newDataPoint PodCPUUsage, maxHistory int32, log logr.Logger, storageLocation string) error {
	switch storageLocation {
	case StorageMemory:
		// Use thread-safe in-memory storage
		key := fmt.Sprintf("%s/%s", namespace, podName)
		log.V(1).Info("Working with in-memory storage", "key", key)
		value, _ := InMemoryMetricsStorage.LoadOrStore(key, []PodCPUUsage{})
		metricsHistory := value.([]PodCPUUsage)
		metricsHistory = append(metricsHistory, newDataPoint)
		if len(metricsHistory) > int(maxHistory) {
			metricsHistory = metricsHistory[len(metricsHistory)-int(maxHistory):]
		}
		InMemoryMetricsStorage.Store(key, metricsHistory)
		log.V(1).Info("Updated in-memory metrics history", "key", key, "historySize", len(metricsHistory))
		return nil
	case StorageAnnotation:
		// Use annotation-based storage
		podKey := client.ObjectKey{Namespace: namespace, Name: podName}
		return retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Fetch the latest version of the pod
			pod := &corev1.Pod{}
			if err := r.Get(ctx, podKey, pod); err != nil {
				log.Error(err, "Failed to fetch pod", "podName", podName)
				return err
			}

			// Initialize or fetch existing metrics history
			var metricsHistory []PodCPUUsage
			if pod.Annotations == nil {
				pod.Annotations = make(map[string]string)
			}
			if historyJSON, exists := pod.Annotations[podMetricsAnnotation]; exists {
				if err := json.Unmarshal([]byte(historyJSON), &metricsHistory); err != nil {
					log.Error(err, "Failed to deserialize existing metrics history", "podName", podName)
					return err
				}
			}

			// Append the new data point and trim history
			metricsHistory = append(metricsHistory, newDataPoint)
			if len(metricsHistory) > int(maxHistory) {
				metricsHistory = metricsHistory[len(metricsHistory)-int(maxHistory):]
			}

			// Serialize updated history
			updatedHistoryJSON, err := json.Marshal(metricsHistory)
			if err != nil {
				log.Error(err, "Failed to serialize updated metrics history", "podName", podName)
				return err
			}

			// Update the pod annotation
			pod.Annotations[podMetricsAnnotation] = string(updatedHistoryJSON)
			if err := r.Update(ctx, pod); err != nil {
				if apierrors.IsConflict(err) {
					log.Info("Conflict detected while updating pod, retrying", "podName", podName)
				} else {
					log.Error(err, "Failed to update pod annotations", "podName", podName)
				}
				return err
			}

			log.V(1).Info("Updated pod metrics history", "podName", podName, "historySize", len(metricsHistory))
			return nil
		})
	default:
		log.Error(fmt.Errorf("unsupported storage location"), "Invalid storage location", "storageLocation", storageLocation)
		return fmt.Errorf("unsupported storage location: %s", storageLocation)
	}
}

func fetchPodMetricsHistory(ctx context.Context, r *MonitorReconciler, pod *corev1.Pod, log logr.Logger, storageLocation string) ([]PodCPUUsage, error) {
	switch storageLocation {
	case StorageMemory:
		// Use thread-safe in-memory storage
		key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		log.V(1).Info("Fetching from in-memory storage", "key", key)
		value, exists := InMemoryMetricsStorage.Load(key)
		if !exists {
			podAge := time.Since(pod.CreationTimestamp.Time)
			log.Info("No in-memory metrics history found", "key", key, "podAge", podAge.String())
			return nil, nil
		}
		metricsHistory := value.([]PodCPUUsage)
		log.V(1).Info("Fetched in-memory metrics history", "key", key, "historySize", len(metricsHistory))
		return metricsHistory, nil
	case StorageAnnotation:
		// Fetch the latest version of the pod
		latestPod := &corev1.Pod{}
		if err := r.Get(ctx, client.ObjectKeyFromObject(pod), latestPod); err != nil {
			log.Error(err, "Failed to fetch latest pod", "podName", pod.Name)
			return nil, err
		}

		// Use annotation-based storage
		metricsHistoryJSON, exists := latestPod.Annotations[podMetricsAnnotation]
		if !exists {
			log.Info("Pod does not have metrics history annotation", "podName", pod.Name)
			return nil, nil
		}

		var metricsHistory []PodCPUUsage
		if err := json.Unmarshal([]byte(metricsHistoryJSON), &metricsHistory); err != nil {
			log.Error(err, "Failed to deserialize metrics history", "podName", pod.Name)
			return nil, err
		}
		log.V(1).Info("Fetched annotation-based metrics history", "podName", pod.Name, "historySize", len(metricsHistory))
		return metricsHistory, nil
	default:
		log.Error(fmt.Errorf("unsupported storage location"), "Invalid storage location", "storageLocation", storageLocation)
		return nil, fmt.Errorf("unsupported storage location: %s", storageLocation)
	}
}

func checkPodMetricsAnnotation(ctx context.Context, r *MonitorReconciler, recycler *recyclertheonlywayecomv1alpha1.Recycler, pod *corev1.Pod, metricsHistory []PodCPUUsage, threshold int32, log logr.Logger) error {
	averageCPU := calculateAverageCPU(metricsHistory, log, pod.Name)

	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		log.Error(err, "Failed to fetch pod", "podName", pod.Name)
		return err
	}

	if averageCPU > float64(threshold) {
		return handleThresholdBreach(ctx, r, recycler, pod, averageCPU, log)
	}

	return handleThresholdRecovery(ctx, r, recycler, pod, averageCPU, log)
}

func calculateAverageCPU(metricsHistory []PodCPUUsage, log logr.Logger, podName string) float64 {
	var totalCPUPercentage float64
	for _, dataPoint := range metricsHistory {
		totalCPUPercentage += dataPoint.CPUPercentage
	}
	averageCPU := totalCPUPercentage / float64(len(metricsHistory))
	log.V(1).Info("Calculated average CPU usage", "podName", podName, "averageCPU", averageCPU)
	return averageCPU
}

func handleThresholdBreach(ctx context.Context, r *MonitorReconciler, recycler *recyclertheonlywayecomv1alpha1.Recycler, pod *corev1.Pod, averageCPU float64, log logr.Logger) error {
	if _, exists := pod.Annotations[cpuBreachTimestampAnnotation]; exists {
		log.V(1).Info("Breach annotation already exists, skipping update", "podName", pod.Name)
		return nil
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
			log.Error(err, "Failed to fetch pod", "podName", pod.Name)
			return err
		}

		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		breachTime := time.Now()
		pod.Annotations[cpuBreachTimestampAnnotation] = breachTime.Format(time.RFC3339)
		if err := r.Update(ctx, pod); err != nil {
			log.Error(err, "Failed to update pod with breach timestamp", "podName", pod.Name)
			return err
		}

		// Calculate termination time based on breach time and delay
		delay := time.Duration(recycler.Spec.RecycleDelaySeconds) * time.Second
		terminationTime := breachTime.Add(delay).Format(time.RFC3339)

		// Calculate pod age
		podAge := time.Since(pod.CreationTimestamp.Time)

		// Write an event to the pod
		r.Recorder.Eventf(pod, corev1.EventTypeWarning, "CPUThresholdBreached",
			"CPU usage threshold breached. Average CPU: %.2f%%", averageCPU)
		// Write an event to the CRD
		r.Recorder.Eventf(recycler, corev1.EventTypeWarning, "CPUThresholdBreached",
			"CPU usage threshold breached for pod %s. Average CPU: %.2f%%", pod.Name, averageCPU)

		log.Info("Breach timestamp annotation added to pod", "podName", pod.Name, "podAge", podAge.String(), "breachTime", breachTime.Format(time.RFC3339), "terminationTime", terminationTime)
		return nil
	})
}

func handleThresholdRecovery(ctx context.Context, r *MonitorReconciler, recycler *recyclertheonlywayecomv1alpha1.Recycler, pod *corev1.Pod, averageCPU float64, log logr.Logger) error {
	if _, exists := pod.Annotations[cpuBreachTimestampAnnotation]; !exists {
		return nil
	}

	log.Info("CPU usage recovered below threshold, removing breach annotation", "podName", pod.Name)

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
			log.Error(err, "Failed to fetch pod", "podName", pod.Name)
			return err
		}

		delete(pod.Annotations, cpuBreachTimestampAnnotation)
		if err := r.Update(ctx, pod); err != nil {
			log.Error(err, "Failed to update pod to remove breach annotation", "podName", pod.Name)
			return err
		}

		// Write an event to the pod
		r.Recorder.Eventf(pod, corev1.EventTypeNormal, "CPUThresholdRecovered",
			"CPU usage recovered below threshold. Average CPU: %.2f%%", averageCPU)
		// Write an event to the CRD
		r.Recorder.Eventf(recycler, corev1.EventTypeNormal, "CPUThresholdRecovered",
			"CPU usage recovered below threshold for pod %s. Average CPU: %.2f%%", pod.Name, averageCPU)

		log.Info("Breach annotation removed from pod", "podName", pod.Name)
		return nil
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register the metrics.k8s.io/v1beta1 API to the scheme
	if err := metricsapi.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("failed to add metrics API to scheme: %w", err)
	}
	r.Recorder = mgr.GetEventRecorder("monitor-controller") // Initialize the EventRecorder
	return ctrl.NewControllerManagedBy(mgr).
		Named("monitor").
		For(&recyclertheonlywayecomv1alpha1.Recycler{}).
		Complete(r)
}
