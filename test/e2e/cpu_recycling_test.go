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

package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
	"github.com/theonlyway/recycler/test/utils"
)

// verifyPodsTerminated checks whether all pods in initialPodNames have been deleted or marked for deletion.
func verifyPodsTerminated(testNamespace string, initialPodNames []string) error {
	terminatedCount := 0
	for _, podName := range initialPodNames {
		cmd := exec.Command("kubectl", "get", "pod", podName,
			"-n", testNamespace, "--ignore-not-found")
		output, _ := utils.Run(cmd)
		if len(output) == 0 {
			terminatedCount++
			continue
		}
		cmd = exec.Command("kubectl", "get", "pod", podName,
			"-n", testNamespace, "-o", "jsonpath={.metadata.deletionTimestamp}")
		deletionTimestamp, _ := utils.Run(cmd)
		if len(deletionTimestamp) > 0 {
			terminatedCount++
		}
	}
	if terminatedCount == len(initialPodNames) {
		return nil
	}
	return fmt.Errorf("%d of %d pods have been terminated", terminatedCount, len(initialPodNames))
}

// verifyRecyclerEvents checks that PodTerminated and CPUThresholdBreached events were recorded on the Recycler CR.
func verifyRecyclerEvents(testNamespace, recyclerName string, initialPodNames []string) error {
	cmd := exec.Command("kubectl", "get", "events",
		"-n", testNamespace,
		"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Recycler", recyclerName),
		"-o", "json",
	)
	output, err := utils.Run(cmd)
	if err != nil {
		return err
	}
	var eventList struct {
		Items []struct {
			Reason string `json:"reason"`
		} `json:"items"`
	}
	if err := json.Unmarshal(output, &eventList); err != nil {
		return fmt.Errorf("failed to parse events: %v", err)
	}
	terminationCount, breachCount := 0, 0
	for _, e := range eventList.Items {
		switch e.Reason {
		case "PodTerminated":
			terminationCount++
		case "CPUThresholdBreached":
			breachCount++
		}
	}
	if terminationCount < len(initialPodNames) {
		return fmt.Errorf("expected %d PodTerminated events, got %d", len(initialPodNames), terminationCount)
	}
	if breachCount < len(initialPodNames) {
		return fmt.Errorf("expected %d CPUThresholdBreached events, got %d", len(initialPodNames), breachCount)
	}
	return nil
}

// verifyPodBreachEvents checks that a CPUThresholdBreached event was recorded on each pod.
func verifyPodBreachEvents(testNamespace string, initialPodNames []string) error {
	for _, podName := range initialPodNames {
		cmd := exec.Command("kubectl", "get", "events",
			"-n", testNamespace,
			"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
			"-o", "json",
		)
		output, err := utils.Run(cmd)
		if err != nil {
			return fmt.Errorf("failed to get events for pod %s: %v", podName, err)
		}
		var eventList struct {
			Items []struct {
				Reason string `json:"reason"`
			} `json:"items"`
		}
		if err := json.Unmarshal(output, &eventList); err != nil {
			return fmt.Errorf("failed to parse events for pod %s: %v", podName, err)
		}
		found := false
		for _, e := range eventList.Items {
			if e.Reason == "CPUThresholdBreached" {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("CPUThresholdBreached event not found for pod %s", podName)
		}
	}
	return nil
}

// verifyNewPodsRunning checks that all original pods have been replaced by new running pods.
func verifyNewPodsRunning(testNamespace string, initialPodNames []string) error {
	cmd := exec.Command("kubectl", "get", "pods",
		"-n", testNamespace, "-l", "app=cpu-stress", "-o", "json")
	output, err := utils.Run(cmd)
	if err != nil {
		return err
	}
	var podList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.Unmarshal(output, &podList); err != nil {
		return fmt.Errorf("failed to parse pod list: %v", err)
	}
	if len(podList.Items) != len(initialPodNames) {
		return fmt.Errorf("expected %d pods but found %d", len(initialPodNames), len(podList.Items))
	}
	for _, newPod := range podList.Items {
		for _, initialPod := range initialPodNames {
			if newPod.Metadata.Name == initialPod {
				return fmt.Errorf("pod %s hasn't been replaced yet", initialPod)
			}
		}
		if newPod.Status.Phase != podStatusRunning {
			return fmt.Errorf("new pod %s not running yet, status: %s", newPod.Metadata.Name, newPod.Status.Phase)
		}
	}
	GinkgoWriter.Printf("All %d pods successfully replaced and running\n", len(podList.Items))
	return nil
}

func cpuRecyclingTest() {
	const testNamespace = "cpu-test"
	const deploymentName = "cpu-stress"
	const recyclerName = "cpu-stress-recycler"
	const labelNamespace = "namespace"
	const labelRecycler = "recycler"
	var err error

	By("creating test namespace")
	cmd := exec.Command("kubectl", "create", "ns", testNamespace)
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("deploying a CPU stress test deployment")
	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = utils.StringReader(cpuStressDeploymentYAML)
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("waiting for CPU stress pod to be running")
	EventuallyWithOffset(1, func() error {
		cmd = exec.Command("kubectl", "get", "pods",
			"-n", testNamespace, "-l", "app=cpu-stress",
			"-o", "jsonpath={.items[0].status.phase}")
		status, err := utils.Run(cmd)
		if err != nil {
			return err
		}
		if string(status) != podStatusRunning {
			return fmt.Errorf("stress pod in %s status", status)
		}
		return nil
	}, 2*time.Minute, 2*time.Second).Should(Succeed())

	By("getting all initial stress pod names")
	cmd = exec.Command("kubectl", "get", "pods",
		"-n", testNamespace, "-l", "app=cpu-stress", "-o", "json")
	podListOutput, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var podList struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	err = json.Unmarshal(podListOutput, &podList)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	initialPodNames := make([]string, 0, len(podList.Items))
	for _, pod := range podList.Items {
		initialPodNames = append(initialPodNames, pod.Metadata.Name)
	}
	ExpectWithOffset(1, initialPodNames).ShouldNot(BeEmpty(), "Expected at least one pod")
	GinkgoWriter.Printf("Initial pods to monitor: %v\n", initialPodNames)

	By("applying the Recycler CR")
	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = utils.StringReader(cpuStressRecyclerYAML)
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("fetching the created Recycler CR to get actual configuration values")
	cmd = exec.Command("kubectl", "get", "recycler", recyclerName,
		"-n", testNamespace, "-o", "json")
	recyclerJSON, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	var recyclerConfig recyclertheonlywayecomv1alpha1.Recycler
	err = json.Unmarshal(recyclerJSON, &recyclerConfig)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	recycleDelaySeconds := recyclerConfig.Spec.RecycleDelaySeconds
	pollingIntervalSeconds := recyclerConfig.Spec.PollingIntervalSeconds
	podMetricsHistory := recyclerConfig.Spec.PodMetricsHistory
	gracePeriodSeconds := recyclerConfig.Spec.GracePeriodSeconds

	GinkgoWriter.Printf("Recycler CR configuration from cluster:\n")
	GinkgoWriter.Printf("  - recycleDelaySeconds: %d\n", recycleDelaySeconds)
	GinkgoWriter.Printf("  - pollingIntervalSeconds: %d\n", pollingIntervalSeconds)
	GinkgoWriter.Printf("  - podMetricsHistory: %d\n", podMetricsHistory)
	GinkgoWriter.Printf("  - gracePeriodSeconds: %d\n", gracePeriodSeconds)

	ExpectWithOffset(1, recycleDelaySeconds).Should(BeNumerically(">", 0))
	ExpectWithOffset(1, pollingIntervalSeconds).Should(BeNumerically(">", 0))
	ExpectWithOffset(1, podMetricsHistory).Should(BeNumerically(">", 0))

	By("waiting for Recycler to have Available status")
	EventuallyWithOffset(1, func() error {
		cmd = exec.Command("kubectl", "get", "recycler", recyclerName,
			"-n", testNamespace,
			"-o", "jsonpath={.status.conditions[?(@.type=='Available')].status}")
		status, err := utils.Run(cmd)
		if err != nil {
			return err
		}
		if string(status) != "True" {
			return fmt.Errorf("recycler not yet available, status: %s", status)
		}
		return nil
	}, 30*time.Second, 2*time.Second).Should(Succeed())

	By("verifying recycler_pod_cpu_utilization_percent is present while pods are running")
	EventuallyWithOffset(1, func() error {
		freshBody, err := utils.FetchControllerMetrics(namespace)
		if err != nil {
			return err
		}
		_, found := utils.MetricValue(freshBody, "recycler_pod_cpu_utilization_percent",
			map[string]string{labelNamespace: testNamespace})
		if !found {
			return fmt.Errorf("recycler_pod_cpu_utilization_percent not found in /metrics for namespace %s", testNamespace)
		}
		return nil
	}, time.Duration(pollingIntervalSeconds*podMetricsHistory)*time.Second+30*time.Second, time.Second).Should(Succeed())

	metricsCollectionTime := time.Duration(pollingIntervalSeconds*podMetricsHistory) * time.Second
	reconcileInterval := 10 * time.Second
	terminationTimeout := metricsCollectionTime +
		time.Duration(recycleDelaySeconds)*time.Second +
		time.Duration(gracePeriodSeconds)*time.Second +
		reconcileInterval +
		60*time.Second

	By(fmt.Sprintf("waiting for all %d pods to be terminated (timeout: %s)", len(initialPodNames), terminationTimeout))
	GinkgoWriter.Printf("Termination timeout breakdown:\n")
	GinkgoWriter.Printf("  - Metrics collection: %s (%ds polling x %d datapoints)\n",
		metricsCollectionTime, pollingIntervalSeconds, podMetricsHistory)
	GinkgoWriter.Printf("  - Recycle delay: %ds\n", recycleDelaySeconds)
	GinkgoWriter.Printf("  - Grace period: %ds\n", gracePeriodSeconds)
	GinkgoWriter.Printf("  - Reconcile interval: %s\n", reconcileInterval)
	GinkgoWriter.Printf("  - Overhead buffer: 1m0s\n")
	GinkgoWriter.Printf("  - Total timeout: %s\n", terminationTimeout)

	defer func() {
		By("capturing operator logs")
		cmd = exec.Command("kubectl", "logs",
			"-l", "control-plane=controller-manager",
			"-n", namespace, "--tail=100", "--all-containers=true")
		logs, err := utils.Run(cmd)
		if err == nil {
			GinkgoWriter.Printf("\n=== Operator Logs ===\n%s\n", string(logs))
		}
	}()

	EventuallyWithOffset(1, func() error {
		return verifyPodsTerminated(testNamespace, initialPodNames)
	}, terminationTimeout, 5*time.Second).Should(Succeed())

	By("capturing all events related to the initial test pods")
	for _, podName := range initialPodNames {
		cmd = exec.Command("kubectl", "get", "events",
			"-n", testNamespace,
			"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
			"-o", "wide",
		)
		podEvents, err := utils.Run(cmd)
		if err == nil {
			GinkgoWriter.Printf("\n=== Events for Pod %s ===\n%s\n", podName, string(podEvents))
		} else {
			GinkgoWriter.Printf("\n=== Failed to get events for pod %s: %v ===\n", podName, err)
		}
	}

	By("capturing all events related to the Recycler CR")
	cmd = exec.Command("kubectl", "get", "events",
		"-n", testNamespace,
		"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Recycler", recyclerName),
		"-o", "wide",
	)
	recyclerEvents, err := utils.Run(cmd)
	if err == nil {
		GinkgoWriter.Printf("\n=== Recycler CR Events ===\n%s\n", string(recyclerEvents))
	} else {
		GinkgoWriter.Printf("\n=== Failed to get Recycler events: %v ===\n", err)
	}

	By("capturing ALL events in the test namespace for debugging")
	cmd = exec.Command("kubectl", "get", "events", "-n", testNamespace, "-o", "wide")
	allEvents, err := utils.Run(cmd)
	if err == nil {
		GinkgoWriter.Printf("\n=== All Events in %s Namespace ===\n%s\n", testNamespace, string(allEvents))
	} else {
		GinkgoWriter.Printf("\n=== Failed to get all events: %v ===\n", err)
	}

	By("verifying CPUThresholdBreached and PodTerminated events were recorded on Recycler CR")
	EventuallyWithOffset(1, func() error {
		return verifyRecyclerEvents(testNamespace, recyclerName, initialPodNames)
	}, 30*time.Second, 2*time.Second).Should(Succeed())

	By("verifying CPUThresholdBreached events were recorded on the pods")
	EventuallyWithOffset(1, func() error {
		return verifyPodBreachEvents(testNamespace, initialPodNames)
	}, 30*time.Second, 2*time.Second).Should(Succeed())

	By("verifying new pods were created by deployment to replace all terminated pods")
	EventuallyWithOffset(1, func() error {
		return verifyNewPodsRunning(testNamespace, initialPodNames)
	}, 2*time.Minute, 5*time.Second).Should(Succeed())

	// recycler_pod_cpu_utilization_percent is deleted immediately when a pod is terminated,
	// so the old pod names should no longer appear in /metrics.
	// recycler_pod_last_recycle_timestamp_seconds intentionally lingers for 5 minutes to allow
	// Prometheus to scrape the final value — we do not assert its disappearance here.
	By("verifying recycler_pod_cpu_utilization_percent no longer reports terminated pod names")
	EventuallyWithOffset(1, func() error {
		freshBody, err := utils.FetchControllerMetrics(namespace)
		if err != nil {
			return err
		}
		for _, podName := range initialPodNames {
			_, found := utils.MetricValue(freshBody, "recycler_pod_cpu_utilization_percent",
				map[string]string{labelNamespace: testNamespace, "recycler_pod": podName})
			if found {
				return fmt.Errorf("recycler_pod_cpu_utilization_percent still present for terminated pod %s", podName)
			}
		}
		return nil
	}, time.Duration(pollingIntervalSeconds)*time.Second+10*time.Second, time.Second).Should(Succeed())

	By("verifying custom Prometheus metrics on the /metrics endpoint")
	metricsBody, err := utils.FetchControllerMetrics(namespace)
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to fetch /metrics from controller")
	GinkgoWriter.Printf("\n=== /metrics excerpt (recycler_* lines) ===\n")
	for _, line := range strings.Split(metricsBody, "\n") {
		if strings.HasPrefix(line, "recycler_") {
			GinkgoWriter.Printf("%s\n", line)
		}
	}

	By("verifying recycler_pod_recycles_total >= number of initial pods")
	recyclesTotal := utils.SumMetricValues(metricsBody, "recycler_pod_recycles_total",
		map[string]string{labelNamespace: testNamespace, labelRecycler: recyclerName})
	ExpectWithOffset(1, recyclesTotal).To(BeNumerically(">=", float64(len(initialPodNames))),
		"expected at least %d recycles, got %.0f", len(initialPodNames), recyclesTotal)

	By("verifying recycler_cpu_threshold_breaches_total >= number of initial pods")
	breachesTotal := utils.SumMetricValues(metricsBody, "recycler_cpu_threshold_breaches_total",
		map[string]string{labelNamespace: testNamespace, labelRecycler: recyclerName})
	ExpectWithOffset(1, breachesTotal).To(BeNumerically(">=", float64(len(initialPodNames))),
		"expected at least %d breach events, got %.0f", len(initialPodNames), breachesTotal)

	By("verifying recycler_pod_last_recycle_timestamp_seconds records each initial pod")
	for _, podName := range initialPodNames {
		ts, tsFound := utils.MetricValue(metricsBody, "recycler_pod_last_recycle_timestamp_seconds",
			map[string]string{labelNamespace: testNamespace, labelRecycler: recyclerName, "pod": podName})
		ExpectWithOffset(1, tsFound).To(BeTrue(),
			"recycler_pod_last_recycle_timestamp_seconds not found for pod %s", podName)
		ExpectWithOffset(1, ts).To(BeNumerically(">", 0),
			"expected non-zero recycle timestamp for pod %s", podName)
	}

	By("cleaning up test resources")
	cmd = exec.Command("kubectl", "delete", "recycler", recyclerName, "-n", testNamespace)
	_, _ = utils.Run(cmd)
	cmd = exec.Command("kubectl", "delete", "deployment", deploymentName, "-n", testNamespace)
	_, _ = utils.Run(cmd)
	cmd = exec.Command("kubectl", "delete", "ns", testNamespace)
	_, _ = utils.Run(cmd)
}
