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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

// mockEventRecorder is a mock implementation of record.EventRecorder for testing
type mockEventRecorder struct{}

func (m *mockEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {}
func (m *mockEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
func (m *mockEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

var _ = Describe("Recycler Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		recycler := &recyclertheonlywayecomv1alpha1.Recycler{}

		BeforeEach(func() {
			By("Creating the target dummy deployment")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "target-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:    "test",
								Image:   "busybox",
								Command: []string{"sleep", "3600"},
							}},
						},
					},
				},
			}
			_ = k8sClient.Create(ctx, deployment) // Ignore error if already exists

			By("creating the custom resource for the Kind Recycler")
			err := k8sClient.Get(ctx, typeNamespacedName, recycler)
			if err != nil && errors.IsNotFound(err) {
				resource := &recyclertheonlywayecomv1alpha1.Recycler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
						ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
							Kind:       "Deployment",
							Name:       "target-deployment",
							APIVersion: "apps/v1",
						},
						AverageCpuUtilizationPercent: 50,
						RecycleDelaySeconds:          300,
						PollingIntervalSeconds:       60,
						PodMetricsHistory:            10,
						GracePeriodSeconds:           30,
						MetricStorageLocation:        "memory",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("Cleanup the specific resource instance Recycler")
			resource := &recyclertheonlywayecomv1alpha1.Recycler{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
				
				// Reconcile to process finalizers
				controllerReconciler := &RecyclerReconciler{
					Client:   k8sClient,
					Scheme:   k8sClient.Scheme(),
					Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
					Recorder: &mockEventRecorder{},
				}
				_, _ = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				
				// Wait for deletion to complete
				Eventually(func() bool {
					err := k8sClient.Get(ctx, typeNamespacedName, resource)
					return errors.IsNotFound(err)
				}, "10s", "100ms").Should(BeTrue())
			}

			By("Cleanup the target deployment")
			deployment := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "target-deployment", Namespace: "default"}, deployment)
			if err == nil {
				Expect(k8sClient.Delete(ctx, deployment)).To(Succeed())
				// Wait for deletion to complete
				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{Name: "target-deployment", Namespace: "default"}, deployment)
					return errors.IsNotFound(err)
				}, "10s", "100ms").Should(BeTrue())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})

		It("should add finalizer to Recycler resource", func() {
			By("Reconciling to add finalizer")
			
			// Wait for resource to be created
			Eventually(func() error {
				recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
				return k8sClient.Get(ctx, typeNamespacedName, recycler)
			}, "5s", "100ms").Should(Succeed())
			
			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Fetch the Recycler resource
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
			err = k8sClient.Get(ctx, typeNamespacedName, recycler)
			Expect(err).NotTo(HaveOccurred())

			// Verify finalizer is added
			Expect(recycler.Finalizers).To(ContainElement("recycler.k8s.io/recycler"))
		})

		It("should update status condition to Available", func() {
			By("Reconciling to update status")
			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Fetch the Recycler resource
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
			err = k8sClient.Get(ctx, typeNamespacedName, recycler)
			Expect(err).NotTo(HaveOccurred())

			// Verify status condition
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, recycler)
				if err != nil {
					return false
				}
				for _, condition := range recycler.Status.Conditions {
					if condition.Type == "Available" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, "10s", "1s").Should(BeTrue())
		})

		It("should return proper requeue duration", func() {
			By("Reconciling and checking requeue")
			
			// Wait for resource to be created
			Eventually(func() error {
				recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
				return k8sClient.Get(ctx, typeNamespacedName, recycler)
			}, "5s", "100ms").Should(Succeed())
			
			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify requeue duration matches polling interval
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{}
			err = k8sClient.Get(ctx, typeNamespacedName, recycler)
			Expect(err).NotTo(HaveOccurred())

			expectedDuration := recycler.Spec.PollingIntervalSeconds
			Expect(result.RequeueAfter.Seconds()).To(Equal(float64(expectedDuration)))
		})

		It("should handle non-existent resource gracefully", func() {
			By("Reconciling a non-existent resource")
			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			nonExistentName := types.NamespacedName{
				Name:      "non-existent",
				Namespace: "default",
			}

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: nonExistentName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
		})

		It("should support annotation storage location", func() {
			By("Creating recycler with annotation storage")
			annotationRecycler := &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "annotation-recycler",
					Namespace: "default",
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "target-deployment",
						APIVersion: "apps/v1",
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        "annotation",
				},
			}
			Expect(k8sClient.Create(ctx, annotationRecycler)).To(Succeed())

			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "annotation-recycler",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter.Seconds()).To(Equal(float64(60)))

			// Cleanup
			Expect(k8sClient.Delete(ctx, annotationRecycler)).To(Succeed())
		})

		It("should handle different polling intervals", func() {
			By("Creating recycler with custom polling interval")
			customRecycler := &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-poll-recycler",
					Namespace: "default",
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "target-deployment",
						APIVersion: "apps/v1",
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       120,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        "memory",
				},
			}
			Expect(k8sClient.Create(ctx, customRecycler)).To(Succeed())

			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			result, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "custom-poll-recycler",
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter.Seconds()).To(Equal(float64(120)))

			// Cleanup
			Expect(k8sClient.Delete(ctx, customRecycler)).To(Succeed())
		})
	})
})
