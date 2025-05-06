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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

var _ = Describe("Recycler Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-recycler"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: resourceName,
			},
		}

		typeNamespaceName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceName, // TODO(user):Modify as needed
		}

		recycler := &recyclertheonlywayecomv1alpha1.Recycler{}

		BeforeEach(func() {
			By("creating the Namespace for the test")
			err := k8sClient.Get(ctx, types.NamespacedName{Name: namespace.Name}, namespace)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, namespace)
				Expect(err).To(Not(HaveOccurred()))
			} else if err != nil && errors.IsAlreadyExists(err) {
				// Namespace already exists, proceed without error
			} else {
				Expect(err).To(Not(HaveOccurred()))
			}

			By("creating the target deployment for the Recycler")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test-app"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test-app"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
								// Removed ProgressDeadlineSeconds from PodTemplateSpec
							},
						},
					},
					// Correctly place ProgressDeadlineSeconds in DeploymentSpec
					ProgressDeadlineSeconds: int32Ptr(30),
				},
			}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, deployment)
				Expect(err).To(Not(HaveOccurred()))
			} else {
				Expect(err).To(Not(HaveOccurred()))
			}

			By("waiting for the target deployment to be ready")
			Eventually(func() error {
				found := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
				if err != nil {
					return err
				}
				if found.Status.ReadyReplicas != *deployment.Spec.Replicas {
					return fmt.Errorf("deployment not ready: %d/%d replicas ready", found.Status.ReadyReplicas, *deployment.Spec.Replicas)
				}
				return nil
			}, 2*time.Minute, time.Second).Should(Succeed())

			By("creating the custom resource for the Kind Recycler")
			err = k8sClient.Get(ctx, typeNamespaceName, recycler)
			if err != nil && errors.IsNotFound(err) {
				resource := &recyclertheonlywayecomv1alpha1.Recycler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace.Name,
					},
					Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
						ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "test-deployment",
						},
						AverageCpuUtilizationPercent: 1,
						RecycleDelaySeconds:          30,
					},
				}
				err = k8sClient.Create(ctx, resource)
				Expect(err).To(Not(HaveOccurred()))
			} else {
				Expect(err).To(Not(HaveOccurred()))
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &recyclertheonlywayecomv1alpha1.Recycler{}
			err := k8sClient.Get(ctx, typeNamespaceName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Recycler")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RecyclerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the Recycler status condition")
			Eventually(func() error {
				found := &recyclertheonlywayecomv1alpha1.Recycler{}
				err := k8sClient.Get(ctx, typeNamespaceName, found)
				if err != nil {
					return err
				}
				if len(found.Status.Conditions) > 0 {
					latestCondition := found.Status.Conditions[len(found.Status.Conditions)-1]
					expectedCondition := metav1.Condition{
						Type:    "Available",
						Status:  metav1.ConditionTrue,
						Reason:  "Monitoring",
						Message: "Recycler is healthy and monitoring the target resource",
					}
					if latestCondition.Type != expectedCondition.Type ||
						latestCondition.Status != expectedCondition.Status ||
						latestCondition.Reason != expectedCondition.Reason ||
						latestCondition.Message != expectedCondition.Message {
						return fmt.Errorf("unexpected status condition: %+v", latestCondition)
					}
				} else {
					return fmt.Errorf("no status conditions found")
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
		})
	})

	Context("Recycler controller test", func() {
		const RecyclerName = "test-recycler"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: RecyclerName,
			},
		}

		typeNamespaceName := types.NamespacedName{
			Name:      RecyclerName,
			Namespace: RecyclerName,
		}

		recycler := &recyclertheonlywayecomv1alpha1.Recycler{}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Get(ctx, types.NamespacedName{Name: namespace.Name}, namespace)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, namespace)
				Expect(err).To(Not(HaveOccurred()))
			} else if err != nil && errors.IsAlreadyExists(err) {
				// Namespace already exists, proceed without error
			} else {
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Creating the target deployment for the Recycler")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test-app"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test-app"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
					// Correctly place ProgressDeadlineSeconds in DeploymentSpec
					ProgressDeadlineSeconds: int32Ptr(30),
				},
			}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, deployment)
				Expect(err).To(Not(HaveOccurred()))
			} else {
				Expect(err).To(Not(HaveOccurred()))
			}

			By("waiting for the target deployment to be ready")
			Eventually(func() error {
				found := &appsv1.Deployment{}
				err := k8sClient.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
				if err != nil {
					return err
				}
				if found.Status.ReadyReplicas != *deployment.Spec.Replicas {
					return fmt.Errorf("deployment not ready: %d/%d replicas ready", found.Status.ReadyReplicas, *deployment.Spec.Replicas)
				}
				return nil
			}, 2*time.Minute, time.Second).Should(Succeed())

			By("Creating the custom resource for the Kind Recycler")
			recycler = &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      RecyclerName,
					Namespace: namespace.Name,
				},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "test-deployment",
					},
					AverageCpuUtilizationPercent: 0,
					RecycleDelaySeconds:          30,
					GracePeriodSeconds:           5,
					PollingIntervalSeconds:       10,
				},
			}
			err = k8sClient.Create(ctx, recycler)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			By("Removing the target deployment for the Recycler")
			deployment := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-deployment", Namespace: namespace.Name}, deployment)
			Expect(err).To(Not(HaveOccurred()))
			Expect(k8sClient.Delete(ctx, deployment)).To(Succeed())

			By("Removing the custom resource for the Kind Recycler")
			found := &recyclertheonlywayecomv1alpha1.Recycler{}
			err = k8sClient.Get(ctx, typeNamespaceName, found)
			Expect(err).To(Not(HaveOccurred()))

			Eventually(func() error {
				return k8sClient.Delete(context.TODO(), found)
			}, 2*time.Minute, time.Second).Should(Succeed())

			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("should successfully reconcile a custom resource for Recycler", func() {
			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &recyclertheonlywayecomv1alpha1.Recycler{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			recyclerReconciler := &RecyclerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := recyclerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if the Recycler status condition is updated")
			Eventually(func() error {
				found := &recyclertheonlywayecomv1alpha1.Recycler{}
				err := k8sClient.Get(ctx, typeNamespaceName, found)
				if err != nil {
					return err
				}
				if len(found.Status.Conditions) > 0 {
					latestCondition := found.Status.Conditions[len(found.Status.Conditions)-1]
					expectedCondition := metav1.Condition{
						Type:    "Available",
						Status:  metav1.ConditionTrue,
						Reason:  "Monitoring",
						Message: "Recycler is healthy and monitoring the target resource",
					}
					// Compare individual fields instead of the entire struct to avoid mismatches.
					if latestCondition.Type != expectedCondition.Type ||
						latestCondition.Status != expectedCondition.Status ||
						latestCondition.Reason != expectedCondition.Reason ||
						latestCondition.Message != expectedCondition.Message {
						return fmt.Errorf("unexpected status condition: %+v", latestCondition)
					}
				} else {
					return fmt.Errorf("no status conditions found")
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})

func int32Ptr(i int32) *int32 {
	return &i
}
