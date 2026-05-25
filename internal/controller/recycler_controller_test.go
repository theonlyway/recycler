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
	"time"

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

// mockEventRecorder is a mock implementation of events.EventRecorder for testing
type mockEventRecorder struct{}

func (m *mockEventRecorder) Eventf(regarding runtime.Object, related runtime.Object, eventtype, reason, action, note string, args ...any) {
}

const (
	targetDeploymentName      = "target-deployment"
	finalizerTestRecyclerName = "finalizer-test-recycler"
	testAppValue              = "test"
	busyboxImage              = "busybox"
)

var _ = Describe("Recycler Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: testNamespace, // TODO(user):Modify as needed
		}
		recycler := &recyclertheonlywayecomv1alpha1.Recycler{}

		BeforeEach(func() {
			By("Creating the target dummy deployment")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      targetDeploymentName,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{appLabelKey: testAppValue},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{appLabelKey: testAppValue},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:    testAppValue,
								Image:   busyboxImage,
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
						Namespace: testNamespace,
					},
					Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
						ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
							Kind:       kindDeployment,
							Name:       targetDeploymentName,
							APIVersion: appsV1APIVersion,
						},
						AverageCpuUtilizationPercent: 50,
						RecycleDelaySeconds:          300,
						PollingIntervalSeconds:       60,
						PodMetricsHistory:            10,
						GracePeriodSeconds:           30,
						MetricStorageLocation:        StorageMemory,
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
			err = k8sClient.Get(ctx, types.NamespacedName{Name: targetDeploymentName, Namespace: testNamespace}, deployment)
			if err == nil {
				Expect(k8sClient.Delete(ctx, deployment)).To(Succeed())
				// Wait for deletion to complete
				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{Name: targetDeploymentName, Namespace: testNamespace}, deployment)
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
				Namespace: testNamespace,
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
					Namespace: testNamespace,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       targetDeploymentName,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        StorageAnnotation,
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
					Namespace: testNamespace,
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
					Namespace: testNamespace,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       targetDeploymentName,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       120,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        StorageMemory,
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
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter.Seconds()).To(Equal(float64(120)))

			// Cleanup
			Expect(k8sClient.Delete(ctx, customRecycler)).To(Succeed())
		})

		It("should handle deployment not found error", func() {
			By("Creating recycler pointing to non-existent deployment")

			nonExistentRecycler := &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-deployment-recycler",
					Namespace: testNamespace,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       "non-existent-deployment",
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        StorageMemory,
				},
			}
			Expect(k8sClient.Create(ctx, nonExistentRecycler)).To(Succeed())

			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "no-deployment-recycler",
					Namespace: testNamespace,
				},
			})
			// Should return error when deployment not found
			Expect(err).To(HaveOccurred())

			// Cleanup
			Expect(k8sClient.Delete(ctx, nonExistentRecycler)).To(Succeed())
		})

		It("should process finalizer operations on deletion", func() {
			By("Creating and deleting a recycler to test finalizer cleanup")

			finalizerRecycler := &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      finalizerTestRecyclerName,
					Namespace: testNamespace,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       targetDeploymentName,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        StorageMemory,
				},
			}
			Expect(k8sClient.Create(ctx, finalizerRecycler)).To(Succeed())

			controllerReconciler := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("controllers").WithName("Recycler"),
				Recorder: &mockEventRecorder{},
			}

			// First reconcile to add finalizer
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      finalizerTestRecyclerName,
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			// Delete the resource
			Expect(k8sClient.Delete(ctx, finalizerRecycler)).To(Succeed())

			// Reconcile again to process finalizer
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      finalizerTestRecyclerName,
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify resource is eventually deleted
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      finalizerTestRecyclerName,
					Namespace: testNamespace,
				}, finalizerRecycler)
				return errors.IsNotFound(err)
			}, "10s", "100ms").Should(BeTrue())
		})
	})

	Context("terminatePods function", func() {
		const (
			terminateDeploymentName = "terminate-test-deployment"
			expiredPodName          = "expired-breach-pod"
			recentPodName           = "recent-breach-pod"
			noAnnotationPodName     = "no-annotation-pod"
			invalidTsPodName        = "invalid-ts-pod"
			terminateTestLabelValue = "terminate-test"
		)

		ctx := context.Background()

		makeTerminateRecycler := func(delaySeconds int32, storage string) *recyclertheonlywayecomv1alpha1.Recycler {
			return &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "terminate-recycler",
					Namespace: testNamespace,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       terminateDeploymentName,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          delaySeconds,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            5,
					GracePeriodSeconds:           0,
					MetricStorageLocation:        storage,
				},
			}
		}

		BeforeEach(func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      terminateDeploymentName,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{appLabelKey: terminateTestLabelValue},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{appLabelKey: terminateTestLabelValue},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  testAppValue,
								Image: busyboxImage,
							}},
						},
					},
				},
			}
			_ = k8sClient.Create(ctx, deployment)
		})

		AfterEach(func() {
			for _, name := range []string{expiredPodName, recentPodName, noAnnotationPodName, invalidTsPodName} {
				pod := &corev1.Pod{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: testNamespace}, pod); err == nil {
					_ = k8sClient.Delete(ctx, pod)
				}
			}
			deployment := &appsv1.Deployment{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: terminateDeploymentName, Namespace: testNamespace}, deployment); err == nil {
				_ = k8sClient.Delete(ctx, deployment)
				Eventually(func() bool {
					return errors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Name: terminateDeploymentName, Namespace: testNamespace}, &appsv1.Deployment{}))
				}, "10s", "100ms").Should(BeTrue())
			}
		})

		createPodWithLabels := func(name string, annotations map[string]string) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        name,
					Namespace:   testNamespace,
					Labels:      map[string]string{appLabelKey: terminateTestLabelValue},
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  testAppValue,
						Image: busyboxImage,
					}},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		}

		newRecyclerReconciler := func() *RecyclerReconciler {
			return &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("test"),
				Recorder: &mockEventRecorder{},
			}
		}

		It("should skip pods without breach timestamp annotation", func() {
			createPodWithLabels(noAnnotationPodName, nil)

			err := terminatePods(ctx, newRecyclerReconciler(), makeTerminateRecycler(5, StorageMemory), ctrl.Log.WithName("test"))
			Expect(err).NotTo(HaveOccurred())

			existingPod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: noAnnotationPodName, Namespace: testNamespace}, existingPod)).To(Succeed())
		})

		It("should skip pods with a recent breach timestamp", func() {
			createPodWithLabels(recentPodName, map[string]string{
				cpuBreachTimestampAnnotation: time.Now().Format(time.RFC3339),
			})

			// Use a 1-hour delay so the recent breach will not have elapsed
			err := terminatePods(ctx, newRecyclerReconciler(), makeTerminateRecycler(3600, StorageMemory), ctrl.Log.WithName("test"))
			Expect(err).NotTo(HaveOccurred())

			existingPod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: recentPodName, Namespace: testNamespace}, existingPod)).To(Succeed())
		})

		It("should terminate pod with expired breach timestamp and clear in-memory storage", func() {
			createPodWithLabels(expiredPodName, map[string]string{
				cpuBreachTimestampAnnotation: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			})

			key := fmt.Sprintf("%s/%s", testNamespace, expiredPodName)
			InMemoryMetricsStorage.Store(key, []PodCPUUsage{{PodName: expiredPodName}})

			// 5-second delay, breach was 10 minutes ago — should trigger termination
			err := terminatePods(ctx, newRecyclerReconciler(), makeTerminateRecycler(5, StorageMemory), ctrl.Log.WithName("test"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Name: expiredPodName, Namespace: testNamespace}, &corev1.Pod{}))
			}, "10s", "100ms").Should(BeTrue())

			_, exists := InMemoryMetricsStorage.Load(key)
			Expect(exists).To(BeFalse())
		})

		It("should terminate pod with expired breach and not clear storage when using annotation storage", func() {
			createPodWithLabels(expiredPodName, map[string]string{
				cpuBreachTimestampAnnotation: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			})

			err := terminatePods(ctx, newRecyclerReconciler(), makeTerminateRecycler(5, StorageAnnotation), ctrl.Log.WithName("test"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Name: expiredPodName, Namespace: testNamespace}, &corev1.Pod{}))
			}, "10s", "100ms").Should(BeTrue())
		})

		It("should skip pods with an invalid breach timestamp", func() {
			createPodWithLabels(invalidTsPodName, map[string]string{
				cpuBreachTimestampAnnotation: "not-a-valid-timestamp",
			})

			err := terminatePods(ctx, newRecyclerReconciler(), makeTerminateRecycler(5, StorageMemory), ctrl.Log.WithName("test"))
			Expect(err).NotTo(HaveOccurred())

			existingPod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: invalidTsPodName, Namespace: testNamespace}, existingPod)).To(Succeed())
		})
	})

	Context("doFinalizerOperationsForRecycler with annotation storage", func() {
		const (
			finalizerAnnotationDeployment = "finalizer-annotation-deployment"
			finalizerAnnotationPod        = "finalizer-annotation-pod"
			finalizerAnnotationLabelValue = "finalizer-annotation-test"
		)

		ctx := context.Background()

		BeforeEach(func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      finalizerAnnotationDeployment,
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{appLabelKey: finalizerAnnotationLabelValue},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{appLabelKey: finalizerAnnotationLabelValue},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  testAppValue,
								Image: busyboxImage,
							}},
						},
					},
				},
			}
			_ = k8sClient.Create(ctx, deployment)
		})

		AfterEach(func() {
			pod := &corev1.Pod{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: finalizerAnnotationPod, Namespace: testNamespace}, pod); err == nil {
				_ = k8sClient.Delete(ctx, pod)
			}
			deployment := &appsv1.Deployment{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: finalizerAnnotationDeployment, Namespace: testNamespace}, deployment); err == nil {
				_ = k8sClient.Delete(ctx, deployment)
				Eventually(func() bool {
					return errors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Name: finalizerAnnotationDeployment, Namespace: testNamespace}, &appsv1.Deployment{}))
				}, "10s", "100ms").Should(BeTrue())
			}
		})

		It("should remove cpu breach and metrics annotations from pods", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      finalizerAnnotationPod,
					Namespace: testNamespace,
					Labels:    map[string]string{appLabelKey: finalizerAnnotationLabelValue},
					Annotations: map[string]string{
						cpuBreachTimestampAnnotation: time.Now().Format(time.RFC3339),
						podMetricsAnnotation:         `[]`,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  testAppValue,
						Image: busyboxImage,
					}},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "finalizer-annotation-recycler",
					Namespace: testNamespace,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       finalizerAnnotationDeployment,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            5,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        StorageAnnotation,
				},
			}

			r := &RecyclerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      ctrl.Log.WithName("test"),
				Recorder: &mockEventRecorder{},
			}

			r.doFinalizerOperationsForRecycler(ctx, recycler)

			updatedPod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: finalizerAnnotationPod, Namespace: testNamespace}, updatedPod)).To(Succeed())
			Expect(updatedPod.Annotations).NotTo(HaveKey(cpuBreachTimestampAnnotation))
			Expect(updatedPod.Annotations).NotTo(HaveKey(podMetricsAnnotation))
		})
	})
})
