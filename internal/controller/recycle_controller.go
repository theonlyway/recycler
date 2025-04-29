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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	recyclerk8siov1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

const recyclerFinalizer = "recycler.k8s.io/recycler"
const (
	typeHealthyCondition   = "Healthy"
	typeUnhealthyCondition = "Unhealthy"
)

// RecycleReconciler reconciles a Recycle object
type RecycleReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Recoder record.EventRecorder
}

// +kubebuilder:rbac:groups=recycler.k8s.io,resources=recycles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=recycler.k8s.io,resources=recycles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=recycler.k8s.io,resources=recycles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Recycle object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *RecycleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Starting Recycler reconciliation")

	recycle := &recyclerk8siov1alpha1.Recycle{}
	err := r.Get(ctx, req.NamespacedName, recycle)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Recycler resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Recycle")
		return ctrl.Result{}, err
	}

	if len(recycle.Status.Conditions) == 0 {
		meta.SetStatusCondition(&recycle.Status.Conditions, metav1.Condition{Type: typeHealthyCondition, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		if err = r.Status().Update(ctx, recycle); err != nil {
			log.Error(err, "unable to update Recycle status")
			return ctrl.Result{}, err
		}

		// Let's re-fetch the recycle Custom Resource after updating the status
		// so that we have the latest state of the resource on the cluster and we will avoid
		// raising the error "the object has been modified, please apply
		// your changes to the latest version and try again" which would re-trigger the reconciliation
		// if we try to update it again in the following operations
		if err := r.Get(ctx, req.NamespacedName, recycle); err != nil {
			log.Error(err, "Failed to re-fetch Recycle")
			return ctrl.Result{}, err
		}
	}

	if !controllerutil.ContainsFinalizer(recycle, recyclerFinalizer) {
		log.Info("Adding finalizer for Recycle")
		if ok := controllerutil.AddFinalizer(recycle, recyclerFinalizer); !ok {
			log.Error(err, "Failed to add finalizer for Recycle custom resource")
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, recycle); err != nil {
			log.Error(err, "Failed to update Recycle custom resource")
			return ctrl.Result{}, err
		}
	}

	log.Info("Ending reconciliation")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecycleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&recyclerk8siov1alpha1.Recycle{}).
		Complete(r)
}
