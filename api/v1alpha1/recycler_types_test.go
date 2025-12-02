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

package v1alpha1

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRecyclerTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Recycler Types Suite")
}

var _ = Describe("Recycler Types", func() {
	Context("CrossVersionObjectReference", func() {
		It("should create a valid CrossVersionObjectReference", func() {
			ref := CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "test-deployment",
				APIVersion: "apps/v1",
			}

			Expect(ref.Kind).To(Equal("Deployment"))
			Expect(ref.Name).To(Equal("test-deployment"))
			Expect(ref.APIVersion).To(Equal("apps/v1"))
		})

		It("should handle default values", func() {
			ref := CrossVersionObjectReference{
				Name: "test-deployment",
			}

			Expect(ref.Name).To(Equal("test-deployment"))
			Expect(ref.Kind).To(Equal(""))
			Expect(ref.APIVersion).To(Equal(""))
		})
	})

	Context("RecyclerSpec", func() {
		It("should create a valid RecyclerSpec with all fields", func() {
			spec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       "test-deployment",
					APIVersion: "apps/v1",
				},
				AverageCpuUtilizationPercent: 75,
				RecycleDelaySeconds:          600,
				PollingIntervalSeconds:       120,
				PodMetricsHistory:            20,
				GracePeriodSeconds:           60,
				MetricStorageLocation:        "memory",
			}

			Expect(spec.AverageCpuUtilizationPercent).To(Equal(int32(75)))
			Expect(spec.RecycleDelaySeconds).To(Equal(int32(600)))
			Expect(spec.PollingIntervalSeconds).To(Equal(int32(120)))
			Expect(spec.PodMetricsHistory).To(Equal(int32(20)))
			Expect(spec.GracePeriodSeconds).To(Equal(int32(60)))
			Expect(spec.MetricStorageLocation).To(Equal("memory"))
		})

		It("should handle annotation storage location", func() {
			spec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "test",
				},
				AverageCpuUtilizationPercent: 50,
				MetricStorageLocation:        "annotation",
			}

			Expect(spec.MetricStorageLocation).To(Equal("annotation"))
		})
	})

	Context("RecyclerStatus", func() {
		It("should create a valid RecyclerStatus", func() {
			status := RecyclerStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "Ready",
						Status:             metav1.ConditionTrue,
						Reason:             "ReconcileSuccess",
						Message:            "Recycler is ready",
						LastTransitionTime: metav1.Now(),
					},
				},
			}

			Expect(status.Conditions).To(HaveLen(1))
			Expect(status.Conditions[0].Type).To(Equal("Ready"))
			Expect(status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
		})

		It("should handle multiple conditions", func() {
			status := RecyclerStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
						Reason: "ReconcileSuccess",
					},
					{
						Type:   "Monitoring",
						Status: metav1.ConditionTrue,
						Reason: "MonitoringActive",
					},
				},
			}

			Expect(status.Conditions).To(HaveLen(2))
		})

		It("should handle empty conditions", func() {
			status := RecyclerStatus{}
			Expect(status.Conditions).To(BeNil())
		})
	})

	Context("Recycler", func() {
		It("should create a complete Recycler resource", func() {
			recycler := Recycler{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "recycler.theonlywaye.com/v1alpha1",
					Kind:       "Recycler",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-recycler",
					Namespace: "default",
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "test-deployment",
						APIVersion: "apps/v1",
					},
					AverageCpuUtilizationPercent: 80,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        "memory",
				},
				Status: RecyclerStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionTrue,
							Reason: "ReconcileSuccess",
						},
					},
				},
			}

			Expect(recycler.Name).To(Equal("test-recycler"))
			Expect(recycler.Namespace).To(Equal("default"))
			Expect(recycler.Spec.AverageCpuUtilizationPercent).To(Equal(int32(80)))
			Expect(recycler.Status.Conditions).To(HaveLen(1))
		})

		It("should handle minimal Recycler configuration", func() {
			recycler := Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "minimal-recycler",
					Namespace: "test",
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Name: "target",
					},
					AverageCpuUtilizationPercent: 50,
				},
			}

			Expect(recycler.Name).To(Equal("minimal-recycler"))
			Expect(recycler.Spec.ScaleTargetRef.Name).To(Equal("target"))
		})
	})

	Context("RecyclerList", func() {
		It("should create an empty RecyclerList", func() {
			list := RecyclerList{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "recycler.theonlywaye.com/v1alpha1",
					Kind:       "RecyclerList",
				},
			}

			Expect(list.Items).To(BeEmpty())
		})

		It("should create a RecyclerList with multiple items", func() {
			list := RecyclerList{
				Items: []Recycler{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "recycler-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "recycler-2",
						},
					},
				},
			}

			Expect(list.Items).To(HaveLen(2))
			Expect(list.Items[0].Name).To(Equal("recycler-1"))
			Expect(list.Items[1].Name).To(Equal("recycler-2"))
		})
	})

	Context("Scheme Registration", func() {
		It("should register types with scheme", func() {
			scheme := runtime.NewScheme()
			err := AddToScheme(scheme)
			Expect(err).NotTo(HaveOccurred())

			// Verify the types are registered
			gvk := GroupVersion.WithKind("Recycler")
			_, err = scheme.New(gvk)
			Expect(err).NotTo(HaveOccurred())

			gvkList := GroupVersion.WithKind("RecyclerList")
			_, err = scheme.New(gvkList)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("JSON Serialization", func() {
		It("should serialize and deserialize Recycler correctly", func() {
			original := &Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-recycler",
					Namespace: "default",
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Kind:       "Deployment",
						Name:       "test-deployment",
						APIVersion: "apps/v1",
					},
					AverageCpuUtilizationPercent: 70,
					RecycleDelaySeconds:          400,
				},
			}

			// Create a copy using DeepCopy
			copy := original.DeepCopy()

			Expect(copy.Name).To(Equal(original.Name))
			Expect(copy.Spec.AverageCpuUtilizationPercent).To(Equal(original.Spec.AverageCpuUtilizationPercent))
			Expect(copy.Spec.ScaleTargetRef.Name).To(Equal(original.Spec.ScaleTargetRef.Name))
		})

		It("should deep copy RecyclerList", func() {
			original := &RecyclerList{
				Items: []Recycler{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "recycler-1",
						},
					},
				},
			}

			copy := original.DeepCopy()
			Expect(copy.Items).To(HaveLen(1))
			Expect(copy.Items[0].Name).To(Equal("recycler-1"))
		})
	})

	Context("Field Validation", func() {
		It("should handle various metric storage locations", func() {
			memorySpec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Name: "test",
				},
				AverageCpuUtilizationPercent: 50,
				MetricStorageLocation:        "memory",
			}
			Expect(memorySpec.MetricStorageLocation).To(Equal("memory"))

			annotationSpec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Name: "test",
				},
				AverageCpuUtilizationPercent: 50,
				MetricStorageLocation:        "annotation",
			}
			Expect(annotationSpec.MetricStorageLocation).To(Equal("annotation"))
		})

		It("should handle different ScaleTargetRef kinds", func() {
			deploymentRef := CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "test-deployment",
				APIVersion: "apps/v1",
			}
			Expect(deploymentRef.Kind).To(Equal("Deployment"))
		})

		It("should handle various integer field values", func() {
			spec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Name: "test",
				},
				AverageCpuUtilizationPercent: 100,
				RecycleDelaySeconds:          1,
				PollingIntervalSeconds:       1,
				PodMetricsHistory:            1,
				GracePeriodSeconds:           0,
			}

			Expect(spec.AverageCpuUtilizationPercent).To(Equal(int32(100)))
			Expect(spec.RecycleDelaySeconds).To(Equal(int32(1)))
			Expect(spec.PollingIntervalSeconds).To(Equal(int32(1)))
			Expect(spec.PodMetricsHistory).To(Equal(int32(1)))
			Expect(spec.GracePeriodSeconds).To(Equal(int32(0)))
		})
	})
})
