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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

const (
	unreachablePromAddress = "http://127.0.0.1:1"
	promTestLabelValue     = "prom-test"
)

var _ = Describe("Prometheus metrics source", func() {
	log := ctrl.Log.WithName("prometheus-test")

	Context("buildPodRegex", func() {
		It("should return an empty string for no pods", func() {
			Expect(buildPodRegex(nil)).To(Equal(""))
			Expect(buildPodRegex([]string{})).To(Equal(""))
		})

		It("should return a single pod name unchanged", func() {
			Expect(buildPodRegex([]string{"web-abc"})).To(Equal("web-abc"))
		})

		It("should join multiple pod names with an alternation", func() {
			Expect(buildPodRegex([]string{"web-abc", "web-def"})).To(Equal("web-abc|web-def"))
		})

		It("should escape regex metacharacters in pod names", func() {
			// A pod name with a dot must be escaped so it cannot match arbitrary characters.
			regex := buildPodRegex([]string{"web-1.2", "api+x"})
			Expect(regex).To(Equal(`web-1\.2|api\+x`))
		})
	})

	Context("renderPrometheusQuery", func() {
		newRecycler := func(query string, history, interval int32) *recyclertheonlywayecomv1alpha1.Recycler {
			return &recyclertheonlywayecomv1alpha1.Recycler{
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					PodMetricsHistory:      history,
					PollingIntervalSeconds: interval,
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						ServerAddress: "http://prometheus:9090",
						Query:         query,
					},
				},
			}
		}

		It("should render the default query when none is supplied", func() {
			recycler := newRecycler("", 5, 30)
			rendered, err := renderPrometheusQuery(recycler, "team-a", "web", []string{"web-1", "web-2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered).To(ContainSubstring(`namespace="team-a"`))
			Expect(rendered).To(ContainSubstring(`pod=~"web-1|web-2"`))
			// WindowSeconds = PodMetricsHistory (5) * PollingIntervalSeconds (30) = 150
			Expect(rendered).To(ContainSubstring("[150s]"))
		})

		It("should render a custom query template with all template fields", func() {
			query := "ns={{.Namespace}} dep={{.Deployment}} pods={{.PodRegex}} win={{.WindowSeconds}}"
			recycler := newRecycler(query, 4, 15)
			rendered, err := renderPrometheusQuery(recycler, "ns1", "dep1", []string{"p1", "p2"})
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered).To(Equal("ns=ns1 dep=dep1 pods=p1|p2 win=60"))
		})

		It("should fall back to a 60s window when history or interval is zero", func() {
			recycler := newRecycler("win={{.WindowSeconds}}", 0, 30)
			rendered, err := renderPrometheusQuery(recycler, "ns", "dep", []string{"p1"})
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered).To(Equal("win=60"))
		})

		It("should return an error for an invalid template", func() {
			recycler := newRecycler("{{.Nope", 5, 30)
			_, err := renderPrometheusQuery(recycler, "ns", "dep", []string{"p1"})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("queryPrometheusCPUUtilization", func() {
		It("should error when the prometheus spec is nil", func() {
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{},
			}
			_, err := queryPrometheusCPUUtilization(ctx, recycler, "ns", "dep", []string{"p1"}, log)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("prometheus configuration is nil"))
		})

		It("should return an empty map when there are no pods (no query issued)", func() {
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						// An unreachable address; it must not be contacted when there are no pods.
						ServerAddress: unreachablePromAddress,
					},
				},
			}
			result, err := queryPrometheusCPUUtilization(ctx, recycler, "ns", "dep", nil, log)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("should error when the prometheus server is unreachable", func() {
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					PodMetricsHistory:      5,
					PollingIntervalSeconds: 30,
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						ServerAddress: unreachablePromAddress,
					},
				},
			}
			_, err := queryPrometheusCPUUtilization(ctx, recycler, "ns", "dep", []string{"p1"}, log)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("prometheus query failed"))
		})
	})

	Context("reconcilePrometheusMetrics", func() {
		It("should return nil when the pod list is empty (no query issued)", func() {
			r := &MonitorReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      log,
				Recorder: &mockEventRecorder{},
			}
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					MetricsSource: MetricsSourcePrometheus,
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						ServerAddress: unreachablePromAddress,
					},
				},
			}
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "dep", Namespace: testNamespace},
			}
			err := reconcilePrometheusMetrics(ctx, r, recycler, deployment, &corev1.PodList{}, log)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should propagate the query error when pods exist and prometheus is unreachable", func() {
			r := &MonitorReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      log,
				Recorder: &mockEventRecorder{},
			}
			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					MetricsSource:          MetricsSourcePrometheus,
					PodMetricsHistory:      5,
					PollingIntervalSeconds: 30,
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						ServerAddress: unreachablePromAddress,
					},
				},
			}
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "dep", Namespace: testNamespace},
			}
			podList := &corev1.PodList{
				Items: []corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "prom-pod", Namespace: testNamespace},
				}},
			}
			err := reconcilePrometheusMetrics(ctx, r, recycler, deployment, podList, log)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("evaluatePodThreshold (shared by both metrics sources)", func() {
		const evalPodName = "eval-threshold-pod"

		ctx := context.Background()

		newReconciler := func() *MonitorReconciler {
			return &MonitorReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      log,
				Recorder: &mockEventRecorder{},
			}
		}

		createPod := func(annotations map[string]string) *corev1.Pod {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        evalPodName,
					Namespace:   testNamespace,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: testAppValue, Image: busyboxImage}},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())
			return pod
		}

		newRecycler := func() *recyclertheonlywayecomv1alpha1.Recycler {
			return &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{Name: "eval-recycler", Namespace: testNamespace},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       "some-deployment",
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 50,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       60,
					PodMetricsHistory:            5,
					GracePeriodSeconds:           30,
					MetricsSource:                MetricsSourcePrometheus,
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						ServerAddress: "http://prometheus:9090",
					},
				},
			}
		}

		AfterEach(func() {
			pod := &corev1.Pod{}
			key := types.NamespacedName{Name: evalPodName, Namespace: testNamespace}
			if err := k8sClient.Get(ctx, key, pod); err == nil {
				_ = k8sClient.Delete(ctx, pod)
				Eventually(func() bool {
					return errors.IsNotFound(k8sClient.Get(ctx, key, &corev1.Pod{}))
				}, "10s", "100ms").Should(BeTrue())
			}
		})

		It("should add the breach annotation when utilization exceeds the threshold", func() {
			createPod(nil)
			pod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: evalPodName, Namespace: testNamespace}, pod)).To(Succeed())

			err := evaluatePodThreshold(ctx, newReconciler(), newRecycler(), pod, 95.0, 50, log)
			Expect(err).NotTo(HaveOccurred())

			updated := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: evalPodName, Namespace: testNamespace}, updated)).To(Succeed())
			Expect(updated.Annotations).To(HaveKey(cpuBreachTimestampAnnotation))
		})

		It("should remove the breach annotation when utilization recovers below the threshold", func() {
			createPod(map[string]string{
				cpuBreachTimestampAnnotation: time.Now().Format(time.RFC3339),
			})
			pod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: evalPodName, Namespace: testNamespace}, pod)).To(Succeed())

			err := evaluatePodThreshold(ctx, newReconciler(), newRecycler(), pod, 10.0, 50, log)
			Expect(err).NotTo(HaveOccurred())

			updated := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: evalPodName, Namespace: testNamespace}, updated)).To(Succeed())
			Expect(updated.Annotations).NotTo(HaveKey(cpuBreachTimestampAnnotation))
		})
	})

	Context("Reconcile with the Prometheus metrics source", func() {
		const (
			promDeploymentName = "prom-monitor-deployment"
			promRecyclerName   = "prom-monitor-recycler"
		)

		ctx := context.Background()
		recyclerKey := types.NamespacedName{Name: promRecyclerName, Namespace: testNamespace}

		BeforeEach(func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: promDeploymentName, Namespace: testNamespace},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{appLabelKey: promTestLabelValue}},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{appLabelKey: promTestLabelValue}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: testAppValue, Image: busyboxImage}},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			// Pods are not scheduled in envtest, so create one explicitly with the matching
			// selector labels so the Prometheus query path is actually exercised.
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prom-test-pod",
					Namespace: testNamespace,
					Labels:    map[string]string{appLabelKey: promTestLabelValue},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: testAppValue, Image: busyboxImage}},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			recycler := &recyclertheonlywayecomv1alpha1.Recycler{
				ObjectMeta: metav1.ObjectMeta{Name: promRecyclerName, Namespace: testNamespace},
				Spec: recyclertheonlywayecomv1alpha1.RecyclerSpec{
					ScaleTargetRef: recyclertheonlywayecomv1alpha1.CrossVersionObjectReference{
						Kind:       kindDeployment,
						Name:       promDeploymentName,
						APIVersion: appsV1APIVersion,
					},
					AverageCpuUtilizationPercent: 75,
					RecycleDelaySeconds:          300,
					PollingIntervalSeconds:       45,
					PodMetricsHistory:            5,
					GracePeriodSeconds:           30,
					MetricStorageLocation:        StorageMemory,
					MetricsSource:                MetricsSourcePrometheus,
					Prometheus: &recyclertheonlywayecomv1alpha1.PrometheusSpec{
						ServerAddress: unreachablePromAddress,
					},
				},
			}
			Expect(k8sClient.Create(ctx, recycler)).To(Succeed())
		})

		AfterEach(func() {
			for _, obj := range []client.Object{
				&recyclertheonlywayecomv1alpha1.Recycler{ObjectMeta: metav1.ObjectMeta{Name: promRecyclerName, Namespace: testNamespace}},
				&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "prom-test-pod", Namespace: testNamespace}},
				&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: promDeploymentName, Namespace: testNamespace}},
			} {
				_ = k8sClient.Delete(ctx, obj)
			}
		})

		It("should take the Prometheus branch and requeue without erroring when the server is unreachable", func() {
			r := &MonitorReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Log:      log,
				Config:   cfg,
				Recorder: &mockEventRecorder{},
			}

			result, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: recyclerKey})
			// The Prometheus query fails (unreachable), but the controller logs and requeues
			// rather than returning an error.
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(45 * time.Second))
		})
	})
})
