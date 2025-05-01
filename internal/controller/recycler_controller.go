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
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

const recyclerFinalizer = "recycler.k8s.io/recycler"
const (
	typeHealthyCondition   = "Available"
	typeUnhealthyCondition = "Unavailable"
)

const recyclerControllerName = "recycler"

// RecyclerReconciler reconciles a Recycler object
type RecyclerReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Recoder record.EventRecorder
}

// +kubebuilder:rbac:groups=recycler.theonlywaye.com,resources=recyclers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=recycler.theonlywaye.com,resources=recyclers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=recycler.theonlywaye.com,resources=recyclers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;update;patch;delete

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
	log := log.FromContext(ctx)
	log.Info("Starting Recycler reconciliation", "controller", recyclerControllerName)

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
			r.doFinalizerOperationsForRecycler(recycler)

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

	return ctrl.Result{}, nil
}

func (r *RecyclerReconciler) doFinalizerOperationsForRecycler(recycler *recyclertheonlywayecomv1alpha1.Recycler) {
	r.Recoder.Event(recycler, "Warning", "Deleting", fmt.Sprintf("Custom resource %s is being deleted from namespace %s", recycler.Name, recycler.Namespace))
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecyclerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recoder = mgr.GetEventRecorderFor("recycler-controller") // Initialize the EventRecorder
	return ctrl.NewControllerManagedBy(mgr).
		Named("recycler").
		For(&recyclertheonlywayecomv1alpha1.Recycler{}).
		Complete(r)
}
