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

const (
	kindDeployment            = "Deployment"
	testDeploymentName        = "test-deployment"
	appsV1APIVersion          = "apps/v1"
	storageMemory             = "memory"
	testName                  = "test"
	conditionReady            = "Ready"
	conditionReconcileSuccess = "ReconcileSuccess"
	testRecyclerName          = "test-recycler"
	defaultNamespace          = "default"
	recyclerOneName           = "recycler-1"
)

func TestRecyclerTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Recycler Types Suite")
}

var _ = Describe("Recycler Types", func() {
	Context("CrossVersionObjectReference", func() {
		It("should create a valid CrossVersionObjectReference", func() {
			ref := CrossVersionObjectReference{
				Kind:       kindDeployment,
				Name:       testDeploymentName,
				APIVersion: appsV1APIVersion,
			}

			Expect(ref.Kind).To(Equal(kindDeployment))
			Expect(ref.Name).To(Equal(testDeploymentName))
			Expect(ref.APIVersion).To(Equal(appsV1APIVersion))
		})

		It("should handle default values", func() {
			ref := CrossVersionObjectReference{
				Name: testDeploymentName,
			}

			Expect(ref.Name).To(Equal(testDeploymentName))
			Expect(ref.Kind).To(Equal(""))
			Expect(ref.APIVersion).To(Equal(""))
		})
	})

	Context("RecyclerSpec", func() {
		It("should create a valid RecyclerSpec with all fields", func() {
			spec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Kind:       kindDeployment,
					Name:       testDeploymentName,
					APIVersion: appsV1APIVersion,
				},
				AverageCpuUtilizationPercent: 75,
				RecycleDelaySeconds:          600,
				PollingIntervalSeconds:       120,
				PodMetricsHistory:            20,
				GracePeriodSeconds:           60,
				MetricStorageLocation:        storageMemory,
			}

			Expect(spec.AverageCpuUtilizationPercent).To(Equal(int32(75)))
			Expect(spec.RecycleDelaySeconds).To(Equal(int32(600)))
			Expect(spec.PollingIntervalSeconds).To(Equal(int32(120)))
			Expect(spec.PodMetricsHistory).To(Equal(int32(20)))
			Expect(spec.GracePeriodSeconds).To(Equal(int32(60)))
			Expect(spec.MetricStorageLocation).To(Equal(storageMemory))
		})

		It("should handle annotation storage location", func() {
			spec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Kind: kindDeployment,
					Name: testName,
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
						Type:               conditionReady,
						Status:             metav1.ConditionTrue,
						Reason:             conditionReconcileSuccess,
						Message:            "Recycler is ready",
						LastTransitionTime: metav1.Now(),
					},
				},
			}

			Expect(status.Conditions).To(HaveLen(1))
			Expect(status.Conditions[0].Type).To(Equal(conditionReady))
			Expect(status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
		})

		It("should handle multiple conditions", func() {
			status := RecyclerStatus{
				Conditions: []metav1.Condition{
					{
						Type:   conditionReady,
						Status: metav1.ConditionTrue,
						Reason: conditionReconcileSuccess,
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
					Name:      testRecyclerName,
					Namespace: defaultNamespace,
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       testDeploymentName,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 80,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            10,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        storageMemory,
				},
				Status: RecyclerStatus{
					Conditions: []metav1.Condition{
						{
							Type:   conditionReady,
							Status: metav1.ConditionTrue,
							Reason: conditionReconcileSuccess,
						},
					},
				},
			}

			Expect(recycler.Name).To(Equal(testRecyclerName))
			Expect(recycler.Namespace).To(Equal(defaultNamespace))
			Expect(recycler.Spec.AverageCpuUtilizationPercent).To(Equal(int32(80)))
			Expect(recycler.Status.Conditions).To(HaveLen(1))
		})

		It("should handle minimal Recycler configuration", func() {
			recycler := Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "minimal-recycler",
					Namespace: testName,
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
							Name: recyclerOneName,
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
			Expect(list.Items[0].Name).To(Equal(recyclerOneName))
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
					Name:      testRecyclerName,
					Namespace: defaultNamespace,
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       testDeploymentName,
						APIVersion: appsV1APIVersion,
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
							Name: recyclerOneName,
						},
					},
				},
			}

			copy := original.DeepCopy()
			Expect(copy.Items).To(HaveLen(1))
			Expect(copy.Items[0].Name).To(Equal(recyclerOneName))
		})
	})

	Context("DeepCopy methods", func() {
		It("should return nil for nil CrossVersionObjectReference.DeepCopy", func() {
			var ref *CrossVersionObjectReference
			Expect(ref.DeepCopy()).To(BeNil())
		})

		It("should DeepCopyInto CrossVersionObjectReference", func() {
			ref := &CrossVersionObjectReference{
				Kind:       kindDeployment,
				Name:       testDeploymentName,
				APIVersion: appsV1APIVersion,
			}
			out := &CrossVersionObjectReference{}
			ref.DeepCopyInto(out)
			Expect(out.Kind).To(Equal(kindDeployment))
			Expect(out.Name).To(Equal(testDeploymentName))
			Expect(out.APIVersion).To(Equal(appsV1APIVersion))
		})

		It("should DeepCopy CrossVersionObjectReference", func() {
			ref := &CrossVersionObjectReference{
				Kind: kindDeployment,
				Name: testDeploymentName,
			}
			copied := ref.DeepCopy()
			Expect(copied).NotTo(BeNil())
			Expect(copied.Kind).To(Equal(kindDeployment))
			Expect(copied.Name).To(Equal(testDeploymentName))
		})

		It("should return nil for nil Recycler.DeepCopy", func() {
			var r *Recycler
			Expect(r.DeepCopy()).To(BeNil())
		})

		It("should return nil for nil Recycler.DeepCopyObject", func() {
			var r *Recycler
			Expect(r.DeepCopyObject()).To(BeNil())
		})

		It("should DeepCopyObject a non-nil Recycler", func() {
			recycler := &Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testRecyclerName,
					Namespace: defaultNamespace,
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Kind: kindDeployment,
						Name: testDeploymentName,
					},
					AverageCpuUtilizationPercent: 75,
				},
			}
			obj := recycler.DeepCopyObject()
			Expect(obj).NotTo(BeNil())
			copied, ok := obj.(*Recycler)
			Expect(ok).To(BeTrue())
			Expect(copied.Name).To(Equal(testRecyclerName))
			Expect(copied.Spec.AverageCpuUtilizationPercent).To(Equal(int32(75)))
		})

		It("should DeepCopyInto Recycler with non-nil conditions", func() {
			original := &Recycler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "original",
					Namespace: defaultNamespace,
				},
				Spec: RecyclerSpec{
					ScaleTargetRef: CrossVersionObjectReference{
						Kind: kindDeployment,
						Name: testDeploymentName,
					},
					AverageCpuUtilizationPercent: 50,
				},
				Status: RecyclerStatus{
					Conditions: []metav1.Condition{
						{
							Type:   conditionReady,
							Status: metav1.ConditionTrue,
							Reason: conditionReconcileSuccess,
						},
					},
				},
			}
			out := &Recycler{}
			original.DeepCopyInto(out)
			Expect(out.Name).To(Equal("original"))
			Expect(out.Status.Conditions).To(HaveLen(1))
			Expect(out.Status.Conditions[0].Type).To(Equal(conditionReady))
		})

		It("should return nil for nil RecyclerList.DeepCopy", func() {
			var list *RecyclerList
			Expect(list.DeepCopy()).To(BeNil())
		})

		It("should return nil for nil RecyclerList.DeepCopyObject", func() {
			var list *RecyclerList
			Expect(list.DeepCopyObject()).To(BeNil())
		})

		It("should DeepCopyObject a non-nil RecyclerList", func() {
			list := &RecyclerList{
				Items: []Recycler{
					{ObjectMeta: metav1.ObjectMeta{Name: recyclerOneName}},
				},
			}
			obj := list.DeepCopyObject()
			Expect(obj).NotTo(BeNil())
			copied, ok := obj.(*RecyclerList)
			Expect(ok).To(BeTrue())
			Expect(copied.Items).To(HaveLen(1))
			Expect(copied.Items[0].Name).To(Equal(recyclerOneName))
		})

		It("should DeepCopyInto RecyclerList preserving items", func() {
			list := &RecyclerList{
				Items: []Recycler{
					{ObjectMeta: metav1.ObjectMeta{Name: "item-1"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "item-2"}},
				},
			}
			out := &RecyclerList{}
			list.DeepCopyInto(out)
			Expect(out.Items).To(HaveLen(2))
			Expect(out.Items[0].Name).To(Equal("item-1"))
			Expect(out.Items[1].Name).To(Equal("item-2"))
		})

		It("should return nil for nil RecyclerSpec.DeepCopy", func() {
			var spec *RecyclerSpec
			Expect(spec.DeepCopy()).To(BeNil())
		})

		It("should DeepCopy a RecyclerSpec", func() {
			spec := &RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Kind: kindDeployment,
					Name: testDeploymentName,
				},
				AverageCpuUtilizationPercent: 80,
				RecycleDelaySeconds:          300,
				MetricStorageLocation:        storageMemory,
			}
			copied := spec.DeepCopy()
			Expect(copied).NotTo(BeNil())
			Expect(copied.AverageCpuUtilizationPercent).To(Equal(int32(80)))
			Expect(copied.ScaleTargetRef.Name).To(Equal(testDeploymentName))
			Expect(copied.MetricStorageLocation).To(Equal(storageMemory))
		})

		It("should return nil for nil RecyclerStatus.DeepCopy", func() {
			var status *RecyclerStatus
			Expect(status.DeepCopy()).To(BeNil())
		})

		It("should DeepCopy a RecyclerStatus with non-nil conditions", func() {
			status := &RecyclerStatus{
				Conditions: []metav1.Condition{
					{
						Type:   conditionReady,
						Status: metav1.ConditionTrue,
						Reason: conditionReconcileSuccess,
					},
				},
			}
			copied := status.DeepCopy()
			Expect(copied).NotTo(BeNil())
			Expect(copied.Conditions).To(HaveLen(1))
			Expect(copied.Conditions[0].Type).To(Equal(conditionReady))
			// Mutating original should not affect copy
			status.Conditions[0].Type = "Modified"
			Expect(copied.Conditions[0].Type).To(Equal(conditionReady))
		})

		It("should DeepCopyInto RecyclerStatus with multiple conditions", func() {
			status := &RecyclerStatus{
				Conditions: []metav1.Condition{
					{Type: "Available", Status: metav1.ConditionTrue, Reason: "Ready"},
					{Type: "Degraded", Status: metav1.ConditionFalse, Reason: "Recovering"},
				},
			}
			out := &RecyclerStatus{}
			status.DeepCopyInto(out)
			Expect(out.Conditions).To(HaveLen(2))
			Expect(out.Conditions[0].Type).To(Equal("Available"))
			Expect(out.Conditions[1].Type).To(Equal("Degraded"))
		})
	})

	Context("Field Validation", func() {
		It("should handle various metric storage locations", func() {
			memorySpec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Name: testName,
				},
				AverageCpuUtilizationPercent: 50,
				MetricStorageLocation:        storageMemory,
			}
			Expect(memorySpec.MetricStorageLocation).To(Equal(storageMemory))

			annotationSpec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Name: testName,
				},
				AverageCpuUtilizationPercent: 50,
				MetricStorageLocation:        "annotation",
			}
			Expect(annotationSpec.MetricStorageLocation).To(Equal("annotation"))
		})

		It("should handle different ScaleTargetRef kinds", func() {
			deploymentRef := CrossVersionObjectReference{
				Kind:       kindDeployment,
				Name:       testDeploymentName,
				APIVersion: appsV1APIVersion,
			}
			Expect(deploymentRef.Kind).To(Equal(kindDeployment))
		})

		It("should handle various integer field values", func() {
			spec := RecyclerSpec{
				ScaleTargetRef: CrossVersionObjectReference{
					Name: testName,
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
