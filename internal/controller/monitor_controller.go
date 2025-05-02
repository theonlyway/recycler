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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

const monitorControllerName = "monitor"

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Recoder record.EventRecorder
}

// PodCPUUsage represents the CPU usage of a pod
type PodCPUUsage struct {
	PodName       string
	CPUUsage      resource.Quantity // Raw CPU usage
	CPULimit      resource.Quantity // CPU limit
	CPUPercentage float64           // Percentage CPU utilization
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
	log := log.FromContext(ctx)

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
		log.Info("Retrieving pods in target deployment", "controller", monitorControllerName, "deployment", recycler.Spec.ScaleTargetRef.Name)
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
		config := ctrl.GetConfigOrDie() // Get the Kubernetes configuration
		metricsClient := resourceclient.NewForConfigOrDie(config).PodMetricses(deployment.Namespace)
		podMetricsList, err := fetchPodMetrics(ctx, metricsClient, deployment.Namespace, deployment.Spec.Selector.MatchLabels, deployment.Spec.Template, log)
		if err != nil {
			log.Error(err, "Failed to fetch metrics for pods in target deployment", "controller", monitorControllerName, "deployment", deploymentKey)
			return ctrl.Result{}, err
		}

		for _, podCPU := range podMetricsList {
			// Update the pod's metrics history annotation
			if err := updatePodMetricsHistory(ctx, r, podCPU.PodName, deployment.Namespace, podCPU, recycler.Spec.PodMetricsHistory, log); err != nil {
				log.Error(err, "Failed to update pod metrics history", "podName", podCPU.PodName)
			}
		}
		return ctrl.Result{RequeueAfter: time.Duration(recycler.Spec.PollingIntervalSeconds) * time.Second}, nil
	default:
		log.Info("Unsupported resource type", "controller", monitorControllerName, "kind", kind)
	}

	return ctrl.Result{RequeueAfter: time.Duration(recycler.Spec.PollingIntervalSeconds) * time.Second}, nil
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
	var podCPUUsages []PodCPUUsage
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

		// Append the pod's CPU utilization to the result list
		podCPUUsages = append(podCPUUsages, PodCPUUsage{
			PodName:       podMetrics.Name,
			CPUUsage:      totalCPUUsage,
			CPULimit:      totalCPULimit,
			CPUPercentage: cpuUtilization,
		})
	}

	return podCPUUsages, nil
}

// UpdatePodMetricsHistory updates the pod's annotation with the latest metrics history
func updatePodMetricsHistory(ctx context.Context, r *MonitorReconciler, podName string, namespace string, newDataPoint PodCPUUsage, maxHistory int32, log logr.Logger) error {
	// Fetch the pod object
	pod := &corev1.Pod{}
	podKey := client.ObjectKey{Namespace: namespace, Name: podName}
	if err := r.Get(ctx, podKey, pod); err != nil {
		log.Error(err, "Failed to fetch pod", "podName", podName)
		return err
	}

	// Initialize or fetch existing metrics history
	var metricsHistory []PodCPUUsage
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	if historyJSON, exists := pod.Annotations["recycler.theonlyway.com/pod-metrics-history"]; exists {
		if err := json.Unmarshal([]byte(historyJSON), &metricsHistory); err != nil {
			log.Error(err, "Failed to deserialize existing metrics history", "podName", podName)
			return err
		}
	}

	// Append the new data point
	metricsHistory = append(metricsHistory, newDataPoint)

	// Trim the history to the maximum allowed size
	if len(metricsHistory) > int(maxHistory) {
		metricsHistory = metricsHistory[len(metricsHistory)-int(maxHistory):]
	}

	// Serialize the updated history back to JSON
	updatedHistoryJSON, err := json.Marshal(metricsHistory)
	if err != nil {
		log.Error(err, "Failed to serialize updated metrics history", "podName", podName)
		return err
	}

	// Update the pod annotation
	pod.Annotations["recycler.theonlyway.com/pod-metrics-history"] = string(updatedHistoryJSON)
	if err := r.Update(ctx, pod); err != nil {
		log.Error(err, "Failed to update pod annotations", "podName", podName)
		return err
	}

	log.Info("Updated pod metrics history", "podName", podName, "historySize", len(metricsHistory))
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register the metrics.k8s.io/v1beta1 API to the scheme
	if err := metricsapi.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("failed to add metrics API to scheme: %w", err)
	}
	r.Recoder = mgr.GetEventRecorderFor("monitor-controller") // Initialize the EventRecorder
	return ctrl.NewControllerManagedBy(mgr).
		Named("monitor").
		For(&recyclertheonlywayecomv1alpha1.Recycler{}).
		Complete(r)
}
