/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

const recyclerFinalizer string = "recycler.k8s.io/recycler"
const (
	typeHealthyCondition         string = "Available"
	typeUnhealthyCondition       string = "Unavailable"
	cpuBreachTimestampAnnotation string = "recycler.theonlyway.com/cpu-breach-timestamp"
	podMetricsAnnotation         string = "recycler.theonlyway.com/pod-metrics-history"
)

const recyclerControllerName string = "recycler"

// RecyclerReconciler reconciles a Recycler object
type RecyclerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder events.EventRecorder
	Log      logr.Logger
}

// +kubebuilder:rbac:groups=recycler.theonlywaye.com,resources=recyclers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=recycler.theonlywaye.com,resources=recyclers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=recycler.theonlywaye.com,resources=recyclers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;update;patch;delete;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Recycler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *RecyclerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("recycler", req.NamespacedName)
	log.V(1).Info("Starting Recycler reconciliation", "controller", recyclerControllerName)

	// Fetch the Recycler instance
	recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
	err := r.Get(ctx, req.NamespacedName, recycler)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Recycler resource not found. Ignoring since object must be deleted", "controller", recyclerControllerName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Recycler", "controller", recyclerControllerName)
		return ctrl.Result{}, err
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(recycler, recyclerFinalizer) {
		log.Info("Adding finalizer for Recycler", "controller", recyclerControllerName)
		controllerutil.AddFinalizer(recycler, recyclerFinalizer)
		if err := r.Update(ctx, recycler); err != nil {
			log.Error(err, "Failed to update Recycler with finalizer", "controller", recyclerControllerName)
			return ctrl.Result{}, err
		}
	}

	// Handle deletion
	if recycler.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(recycler, recyclerFinalizer) {
			log.Info("Performing finalizer operations", "controller", recyclerControllerName)

			// Perform finalizer operations
			r.doFinalizerOperationsForRecycler(ctx, recycler)

			// Remove finalizer
			controllerutil.RemoveFinalizer(recycler, recyclerFinalizer)
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				latestRecycler := &recyclertheonlywayecomv1alpha1.Recycler{}
				if err := r.Get(ctx, req.NamespacedName, latestRecycler); err != nil {
					return err
				}
				controllerutil.RemoveFinalizer(latestRecycler, recyclerFinalizer)
				return r.Update(ctx, latestRecycler)
			})
			if err != nil {
				log.Error(err, "Failed to remove finalizer", "controller", recyclerControllerName)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Update status condition
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latestRecycler := &recyclertheonlywayecomv1alpha1.Recycler{}
		if err := r.Get(ctx, req.NamespacedName, latestRecycler); err != nil {
			return err
		}
		meta.SetStatusCondition(&latestRecycler.Status.Conditions, metav1.Condition{
			Type:    typeHealthyCondition,
			Status:  metav1.ConditionTrue,
			Reason:  "Monitoring",
			Message: "Recycler is healthy and monitoring the target resource",
		})
		return r.Status().Update(ctx, latestRecycler)
	})
	if err != nil {
		log.Error(err, "Failed to update Recycler status", "controller", recyclerControllerName)
		return ctrl.Result{}, err
	}

	if err := terminatePods(ctx, r, recycler, log); err != nil {
		log.Error(err, "Failed to terminate pods", "controller", recyclerControllerName)
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Duration(recycler.Spec.PollingIntervalSeconds) * time.Second}, nil
}

func terminatePods(ctx context.Context, r *RecyclerReconciler, recycler *recyclertheonlywayecomv1alpha1.Recycler, log logr.Logger) error {
	// Fetch the target deployment using ScaleTargetRef
	deployment := &appsv1.Deployment{}
	deploymentKey := client.ObjectKey{
		Namespace: recycler.Namespace,
		Name:      recycler.Spec.ScaleTargetRef.Name,
	}
	if err := r.Get(ctx, deploymentKey, deployment); err != nil {
		log.Error(err, "Failed to fetch target deployment", "controller", recyclerControllerName)
		return err
	}

	// List all pods in the target deployment
	podList := &corev1.PodList{}
	listOptions := []client.ListOption{
		client.InNamespace(deployment.Namespace),
		client.MatchingLabels(deployment.Spec.Selector.MatchLabels),
	}
	if err := r.List(ctx, podList, listOptions...); err != nil {
		log.Error(err, "Failed to list pods for termination", "controller", recyclerControllerName)
		return err
	}

	for _, pod := range podList.Items {
		// Check for the breach timestamp annotation
		breachTimestamp, exists := pod.Annotations[cpuBreachTimestampAnnotation]
		if !exists {
			log.V(1).Info("Pod does not have breach timestamp annotation, skipping", "podName", pod.Name)
			continue
		}

		// Parse the breach timestamp
		breachTime, err := time.Parse(time.RFC3339, breachTimestamp)
		if err != nil {
			log.Error(err, "Failed to parse breach timestamp", "podName", pod.Name, "breachTimestamp", breachTimestamp)
			continue
		}

		// Calculate the time elapsed since the breach
		elapsed := time.Since(breachTime)
		delay := time.Duration(recycler.Spec.RecycleDelaySeconds) * time.Second

		// Calculate pod age
		podAge := time.Since(pod.CreationTimestamp.Time)

		if elapsed >= delay {
			log.Info("Terminating pod due to CPU threshold breach",
				"podName", pod.Name,
				"podAge", podAge.String(),
				"breachTimestamp", breachTimestamp,
				"elapsed", elapsed,
				"delay", delay)

			// Set grace period for pod termination
			gracePeriod := int64(recycler.Spec.GracePeriodSeconds)
			deleteOptions := &client.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
			}

			// Terminate the pod
			if err := r.Delete(ctx, &pod, deleteOptions); err != nil {
				log.Error(err, "Failed to delete pod", "podName", pod.Name)
			} else {
				// Check if in-memory storage is being used
				if recycler.Spec.MetricStorageLocation == "memory" {
					// Remove the pod's entry from in-memory storage
					key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
					InMemoryMetricsStorage.Delete(key) // Access exported variable
					log.V(1).Info("Removed pod entry from in-memory storage", "podName", pod.Name, "key", key)
				}

				// Write an event to the CRD
				r.Recorder.Event(recycler, corev1.EventTypeNormal, "PodTerminated", fmt.Sprintf("Pod %s terminated due to CPU threshold breach", pod.Name))
			}
		} else {
			terminationTime := breachTime.Add(delay)
			log.V(1).Info("Pod not ready for termination yet",
				"podName", pod.Name,
				"breachTimestamp", breachTimestamp,
				"elapsed", elapsed,
				"delay", delay,
				"terminationTime", terminationTime)
		}
	}

	return nil
}

func (r *RecyclerReconciler) doFinalizerOperationsForRecycler(ctx context.Context, recycler *recyclertheonlywayecomv1alpha1.Recycler) {
	r.Recorder.Event(recycler, "Warning", "Deleting", fmt.Sprintf("Custom resource %s is being deleted from namespace %s", recycler.Name, recycler.Namespace))

	// Fetch the target deployment using ScaleTargetRef
	deployment := &appsv1.Deployment{}
	deploymentKey := client.ObjectKey{
		Namespace: recycler.Namespace,
		Name:      recycler.Spec.ScaleTargetRef.Name,
	}
	if err := r.Get(ctx, deploymentKey, deployment); err != nil {
		r.Recorder.Event(recycler, "Warning", "FinalizerError", fmt.Sprintf("Failed to fetch deployment: %v", err))
		return
	}

	// List all pods in the target deployment
	podList := &corev1.PodList{}
	listOptions := []client.ListOption{
		client.InNamespace(deployment.Namespace),
		client.MatchingLabels(deployment.Spec.Selector.MatchLabels),
	}
	if err := r.List(ctx, podList, listOptions...); err != nil {
		r.Recorder.Event(recycler, "Warning", "FinalizerError", fmt.Sprintf("Failed to list pods: %v", err))
		return
	}

	// Handle cleanup based on MetricStorageLocation
	switch recycler.Spec.MetricStorageLocation {
	case "annotation":
		// Remove annotations from each pod
		for _, pod := range podList.Items {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				// Fetch the latest version of the pod
				latestPod := &corev1.Pod{}
				if err := r.Get(ctx, client.ObjectKeyFromObject(&pod), latestPod); err != nil {
					return err
				}

				// Remove annotations
				if latestPod.Annotations != nil {
					delete(latestPod.Annotations, cpuBreachTimestampAnnotation)
					delete(latestPod.Annotations, podMetricsAnnotation)
				}

				// Update the pod
				return r.Update(ctx, latestPod)
			})

			if retryErr != nil {
				r.Recorder.Event(recycler, "Warning", "FinalizerError", fmt.Sprintf("Failed to remove annotations from pod %s: %v", pod.Name, retryErr))
			} else {
				r.Recorder.Event(recycler, "Normal", "AnnotationsRemoved", fmt.Sprintf("Removed annotations from pod %s", pod.Name))
			}
		}
	case "memory":
		// Clear in-memory metrics storage for each pod
		for _, pod := range podList.Items {
			key := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
			InMemoryMetricsStorage.Delete(key)
			r.Log.V(1).Info("Cleared in-memory metrics storage for pod", "podName", pod.Name, "key", key)
		}
	default:
		r.Log.Error(fmt.Errorf("unsupported storage location"), "Invalid MetricStorageLocation", "MetricStorageLocation", recycler.Spec.MetricStorageLocation)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecyclerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorder("recycler-controller") // Initialize the EventRecorder
	return ctrl.NewControllerManagedBy(mgr).
		Named("recycler").
		For(&recyclertheonlywayecomv1alpha1.Recycler{}).
		Complete(r)
}
