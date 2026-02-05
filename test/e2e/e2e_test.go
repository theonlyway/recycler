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
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
	"github.com/theonlyway/recycler/test/utils"
)

//go:embed testdata/cpu-stress-deployment.yaml
var cpuStressDeploymentYAML string

//go:embed testdata/cpu-stress-recycler.yaml
var cpuStressRecyclerYAML string

const namespace = "recycler-system"
const podStatusRunning = "Running"

var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {

		By("installing prometheus operator")
		Expect(utils.InstallPrometheusOperator()).To(Succeed())

		By("installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("installing metrics-server")
		Expect(utils.InstallMetricsServer()).To(Succeed())

		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {

		By("uninstalling the Prometheus manager bundle")
		utils.UninstallPrometheusOperator()

		By("uninstalling the cert-manager bundle")
		utils.UninstallCertManager()

		By("uninstalling metrics-server")
		utils.UninstallMetricsServer()

		By("removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	Context("Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error

			// projectimage stores the name of the image used in the example
			var projectimage = "ghcr.io/theonlyway/recycler:latest"
			var platform = "linux/amd64"

			By("building the manager(Operator) image")
			cmd := exec.Command(
				"make", "docker-buildx",
				fmt.Sprintf("IMG=%s", projectimage),
				fmt.Sprintf("PLATFORM=%s", platform),
				"BUILD_MODE=load",
			)
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			err = utils.LoadImageToKindClusterWithName(projectimage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy-debug", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("listing all resources in the recycler-system namespace")
			cmd = exec.Command("kubectl", "get", "all", "-n", namespace)
			namespaceOutput, err := utils.Run(cmd)
			_, _ = fmt.Fprintf(GinkgoWriter, "Resources in %s namespace:\n%s\n", namespace, string(namespaceOutput))
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != podStatusRunning {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}

			// Capture diagnostic information if the controller fails to start
			captureDebugInfo := func() {
				if controllerPodName == "" {
					return // No pod was found, skip diagnostics
				}

				By("capturing deployment status for debugging")
				cmd = exec.Command("kubectl", "get", "deployment",
					"-n", namespace,
					"-o", "wide",
				)
				deploymentStatus, err := utils.Run(cmd)
				if err == nil {
					GinkgoWriter.Printf("\n=== Deployment Status ===\n%s\n", string(deploymentStatus))
				}

				By("capturing pod details for debugging")
				cmd = exec.Command("kubectl", "describe", "pod", controllerPodName,
					"-n", namespace,
				)
				podDetails, err := utils.Run(cmd)
				if err == nil {
					GinkgoWriter.Printf("\n=== Pod Details ===\n%s\n", string(podDetails))
				}

				By("capturing events in namespace for debugging")
				cmd = exec.Command("kubectl", "get", "events",
					"-n", namespace,
					"--sort-by=.lastTimestamp",
				)
				events, err := utils.Run(cmd)
				if err == nil {
					GinkgoWriter.Printf("\n=== Namespace Events ===\n%s\n", string(events))
				}

				By("capturing pod logs if available")
				cmd = exec.Command("kubectl", "logs", controllerPodName,
					"-n", namespace,
					"--all-containers=true",
					"--ignore-errors",
				)
				logs, err := utils.Run(cmd)
				if err == nil && len(logs) > 0 {
					GinkgoWriter.Printf("\n=== Pod Logs ===\n%s\n", string(logs))
				}
			}

			// Try to verify controller, capture debug info only on failure
			func() {
				defer func() {
					if r := recover(); r != nil {
						captureDebugInfo()
						panic(r) // Re-panic to let Ginkgo handle it
					}
				}()
				EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())
			}()

		})

		It("should terminate pod when CPU threshold is exceeded", func() {
			const testNamespace = "cpu-test"
			const deploymentName = "cpu-stress"
			const recyclerName = "cpu-stress-recycler"
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
			verifyStressPodRunning := func() error {
				cmd = exec.Command("kubectl", "get", "pods",
					"-n", testNamespace,
					"-l", "app=cpu-stress",
					"-o", "jsonpath={.items[0].status.phase}",
				)
				status, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				if string(status) != podStatusRunning {
					return fmt.Errorf("stress pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyStressPodRunning, 2*time.Minute, 2*time.Second).Should(Succeed())

			// Get all initial pod names
			var initialPodNames []string
			By("getting all initial stress pod names")
			cmd = exec.Command("kubectl", "get", "pods",
				"-n", testNamespace,
				"-l", "app=cpu-stress",
				"-o", "jsonpath={.items[*].metadata.name}",
			)
			podNameOutput, err := utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			initialPodNames = utils.GetNonEmptyLines(string(podNameOutput))
			ExpectWithOffset(1, len(initialPodNames)).Should(BeNumerically(">=", 1), "Expected at least one pod")
			GinkgoWriter.Printf("Initial pods to monitor: %v\n", initialPodNames)
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.StringReader(cpuStressRecyclerYAML)
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("fetching the created Recycler CR to get actual configuration values")
			cmd = exec.Command("kubectl", "get", "recycler", recyclerName,
				"-n", testNamespace,
				"-o", "json",
			)
			recyclerJSON, err := utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Parse the actual Recycler CR from the cluster
			var recyclerConfig recyclertheonlywayecomv1alpha1.Recycler
			err = json.Unmarshal(recyclerJSON, &recyclerConfig)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Extract configuration values from the live CR
			recycleDelaySeconds := recyclerConfig.Spec.RecycleDelaySeconds
			pollingIntervalSeconds := recyclerConfig.Spec.PollingIntervalSeconds
			podMetricsHistory := recyclerConfig.Spec.PodMetricsHistory
			gracePeriodSeconds := recyclerConfig.Spec.GracePeriodSeconds

			// Log the actual configuration values being used
			GinkgoWriter.Printf("Recycler CR configuration from cluster:\n")
			GinkgoWriter.Printf("  - recycleDelaySeconds: %d\n", recycleDelaySeconds)
			GinkgoWriter.Printf("  - pollingIntervalSeconds: %d\n", pollingIntervalSeconds)
			GinkgoWriter.Printf("  - podMetricsHistory: %d\n", podMetricsHistory)
			GinkgoWriter.Printf("  - gracePeriodSeconds: %d\n", gracePeriodSeconds)

			// Validate that values are valid
			ExpectWithOffset(1, recycleDelaySeconds).Should(BeNumerically(">", 0),
				"recycleDelaySeconds should be greater than 0")
			ExpectWithOffset(1, pollingIntervalSeconds).Should(BeNumerically(">", 0),
				"pollingIntervalSeconds should be greater than 0")
			ExpectWithOffset(1, podMetricsHistory).Should(BeNumerically(">", 0),
				"podMetricsHistory should be greater than 0")

			By("waiting for Recycler to have Available status")
			verifyRecyclerHealthy := func() error {
				cmd = exec.Command("kubectl", "get", "recycler",
					recyclerName,
					"-n", testNamespace,
					"-o", "jsonpath={.status.conditions[?(@.type=='Available')].status}",
				)
				status, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				if string(status) != "True" {
					return fmt.Errorf("recycler not yet available, status: %s", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyRecyclerHealthy, 30*time.Second, 2*time.Second).Should(Succeed())

			By("Pod termination verification check for all pods")
			verifyPodTerminated := func() error {
				// Check each initial pod to see if it has been terminated
				terminatedCount := 0
				for _, podName := range initialPodNames {
					// Check if the original pod still exists
					cmd = exec.Command("kubectl", "get", "pod",
						podName,
						"-n", testNamespace,
						"--ignore-not-found",
					)
					output, _ := utils.Run(cmd)

					// If the pod no longer exists, it was terminated
					if len(output) == 0 {
						terminatedCount++
						continue
					}

					// Check if pod has deletion timestamp
					cmd = exec.Command("kubectl", "get", "pod",
						podName,
						"-n", testNamespace,
						"-o", "jsonpath={.metadata.deletionTimestamp}",
					)
					deletionTimestamp, _ := utils.Run(cmd)
					if len(deletionTimestamp) > 0 {
						terminatedCount++
					}
				}

				// All initial pods should be terminated
				if terminatedCount == len(initialPodNames) {
					return nil
				}

				return fmt.Errorf("%d of %d pods have been terminated", terminatedCount, len(initialPodNames))
			}
			// Calculate timeout: time to collect enough metrics + recycle delay + grace period + reconcile + buffer
			metricsCollectionTime := time.Duration(pollingIntervalSeconds*podMetricsHistory) * time.Second
			reconcileInterval := 10 * time.Second // Recycler controller reconciles every ~10s
			terminationTimeout := metricsCollectionTime +
				time.Duration(recycleDelaySeconds)*time.Second +
				time.Duration(gracePeriodSeconds)*time.Second +
				reconcileInterval + // Wait for reconciliation cycle after termination time
				60*time.Second // buffer for overhead and Kubernetes operations

			By(fmt.Sprintf("waiting for all %d pods to be terminated due to high CPU usage (timeout: %s)", len(initialPodNames), terminationTimeout))
			GinkgoWriter.Printf("Termination timeout breakdown:\n")
			GinkgoWriter.Printf("  - Metrics collection: %s (%ds polling × %d datapoints)\n",
				metricsCollectionTime, pollingIntervalSeconds, podMetricsHistory)
			GinkgoWriter.Printf("  - Recycle delay: %ds\n", recycleDelaySeconds)
			GinkgoWriter.Printf("  - Grace period: %ds\n", gracePeriodSeconds)
			GinkgoWriter.Printf("  - Reconcile interval: %s (wait for next cycle)\n", reconcileInterval)
			GinkgoWriter.Printf("  - Overhead buffer: 1m0s\n")
			GinkgoWriter.Printf("  - Total timeout: %s\n", terminationTimeout)

			// Capture operator logs for debugging
			defer func() {
				By("capturing operator logs")
				cmd = exec.Command("kubectl", "logs",
					"-l", "control-plane=controller-manager",
					"-n", namespace,
					"--tail=100",
					"--all-containers=true",
				)
				logs, err := utils.Run(cmd)
				if err == nil {
					GinkgoWriter.Printf("\n=== Operator Logs ===\n%s\n", string(logs))
				}
			}()

			EventuallyWithOffset(1, verifyPodTerminated, terminationTimeout, 5*time.Second).Should(Succeed())

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
			cmd = exec.Command("kubectl", "get", "events",
				"-n", testNamespace,
				"-o", "wide",
			)
			allEvents, err := utils.Run(cmd)
			if err == nil {
				GinkgoWriter.Printf("\n=== All Events in %s Namespace ===\n%s\n", testNamespace, string(allEvents))
			} else {
				GinkgoWriter.Printf("\n=== Failed to get all events: %v ===\n", err)
			}

			By("verifying PodTerminated events were recorded on Recycler CR for all pods")
			verifyTerminationEvent := func() error {
				cmd = exec.Command("kubectl", "get", "events",
					"-n", testNamespace,
					"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Recycler", recyclerName),
					"-o", "json",
				)
				events, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				// Count PodTerminated events
				var eventList struct {
					Items []struct {
						Reason  string `json:"reason"`
						Message string `json:"message"`
					} `json:"items"`
				}
				if err := json.Unmarshal(events, &eventList); err != nil {
					return fmt.Errorf("failed to parse events: %v", err)
				}

				terminationEventCount := 0
				for _, event := range eventList.Items {
					if event.Reason == "PodTerminated" {
						terminationEventCount++
					}
				}

				if terminationEventCount < len(initialPodNames) {
					return fmt.Errorf("expected %d PodTerminated events, but found %d", len(initialPodNames), terminationEventCount)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyTerminationEvent, 30*time.Second, 2*time.Second).Should(Succeed())

			By("verifying new pods were created by deployment to replace all terminated pods")
			verifyNewPodCreated := func() error {
				cmd = exec.Command("kubectl", "get", "pods",
					"-n", testNamespace,
					"-l", "app=cpu-stress",
					"-o", "jsonpath={.items[*].metadata.name}",
				)
				newPodNamesOutput, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				newPodNames := utils.GetNonEmptyLines(string(newPodNamesOutput))

				// Should have the same number of pods as initially
				if len(newPodNames) != len(initialPodNames) {
					return fmt.Errorf("expected %d pods but found %d", len(initialPodNames), len(newPodNames))
				}

				// Check that all pods have changed (none match initial names)
				for _, newPod := range newPodNames {
					for _, initialPod := range initialPodNames {
						if newPod == initialPod {
							return fmt.Errorf("pod %s hasn't been replaced yet", initialPod)
						}
					}

					// Verify each new pod is running
					cmd = exec.Command("kubectl", "get", "pod",
						newPod,
						"-n", testNamespace,
						"-o", "jsonpath={.status.phase}",
					)
					status, err := utils.Run(cmd)
					if err != nil {
						return err
					}
					if string(status) != podStatusRunning {
						return fmt.Errorf("new pod %s not running yet, status: %s", newPod, status)
					}
				}

				GinkgoWriter.Printf("All %d pods successfully replaced and running\n", len(newPodNames))
				return nil
			}
			EventuallyWithOffset(1, verifyNewPodCreated, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("cleaning up test resources")
			cmd = exec.Command("kubectl", "delete", "recycler", recyclerName, "-n", testNamespace)
			_, _ = utils.Run(cmd)

			cmd = exec.Command("kubectl", "delete", "deployment", deploymentName, "-n", testNamespace)
			_, _ = utils.Run(cmd)

			cmd = exec.Command("kubectl", "delete", "ns", testNamespace)
			_, _ = utils.Run(cmd)
		})
	})
})
