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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

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
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &recyclertheonlywayecomv1alpha1.Recycler{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
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
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
