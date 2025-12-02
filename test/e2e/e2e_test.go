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
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
	"github.com/theonlyway/recycler/test/utils"
)

//go:embed testdata/cpu-stress-deployment.yaml
var cpuStressDeploymentYAML string

//go:embed testdata/cpu-stress-recycler.yaml
var cpuStressRecyclerYAML string

// parseRecyclerYAML extracts the spec values from the embedded YAML
func parseRecyclerYAML(yamlContent string) (*recyclertheonlywayecomv1alpha1.Recycler, error) {
	var recycler recyclertheonlywayecomv1alpha1.Recycler
	if err := yaml.Unmarshal([]byte(yamlContent), &recycler); err != nil {
		return nil, err
	}
	return &recycler, nil
}

const namespace = "recycler-system"

var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {
		By("installing prometheus operator")
		Expect(utils.InstallPrometheusOperator()).To(Succeed())

		By("installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {
		By("uninstalling the Prometheus manager bundle")
		utils.UninstallPrometheusOperator()

		By("uninstalling the cert-manager bundle")
		utils.UninstallCertManager()

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
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
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
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

		})

		It("should terminate pod when CPU threshold is exceeded", func() {
			const testNamespace = "cpu-test"
			const deploymentName = "cpu-stress"
			const recyclerName = "cpu-stress-recycler"
			var err error

			// Parse the recycler configuration from the YAML
			recyclerConfig, err := parseRecyclerYAML(cpuStressRecyclerYAML)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// Extract configuration values
			recycleDelaySeconds := recyclerConfig.Spec.RecycleDelaySeconds
			pollingIntervalSeconds := recyclerConfig.Spec.PollingIntervalSeconds
			podMetricsHistory := recyclerConfig.Spec.PodMetricsHistory
			gracePeriodSeconds := recyclerConfig.Spec.GracePeriodSeconds

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
				if string(status) != "Running" {
					return fmt.Errorf("stress pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyStressPodRunning, 2*time.Minute, 2*time.Second).Should(Succeed())

			// Get the initial pod name
			var initialPodName string
			By("getting initial stress pod name")
			cmd = exec.Command("kubectl", "get", "pods",
				"-n", testNamespace,
				"-l", "app=cpu-stress",
				"-o", "jsonpath={.items[0].metadata.name}",
			)
			podNameOutput, err := utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			initialPodName = string(podNameOutput)
			ExpectWithOffset(1, initialPodName).ShouldNot(BeEmpty())

			By("creating a Recycler CR to monitor the CPU stress deployment")
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = utils.StringReader(cpuStressRecyclerYAML)
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

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

			By("waiting for the pod to be terminated due to high CPU usage")
			verifyPodTerminated := func() error {
				// Check if the original pod still exists
				cmd = exec.Command("kubectl", "get", "pod",
					initialPodName,
					"-n", testNamespace,
					"--ignore-not-found",
				)
				output, _ := utils.Run(cmd)

				// If the pod no longer exists, it was terminated
				if len(output) == 0 {
					return nil
				}

				// Check if pod has deletion timestamp
				cmd = exec.Command("kubectl", "get", "pod",
					initialPodName,
					"-n", testNamespace,
					"-o", "jsonpath={.metadata.deletionTimestamp}",
				)
				deletionTimestamp, _ := utils.Run(cmd)
				if len(deletionTimestamp) > 0 {
					return nil
				}

				return fmt.Errorf("pod %s has not been terminated yet", initialPodName)
			}
			// Calculate timeout: time to collect enough metrics + recycle delay + grace period + buffer
			// Metrics collection: pollingIntervalSeconds * podMetricsHistory
			metricsCollectionTime := time.Duration(pollingIntervalSeconds*podMetricsHistory) * time.Second
			terminationTimeout := metricsCollectionTime +
				time.Duration(recycleDelaySeconds)*time.Second +
				time.Duration(gracePeriodSeconds)*time.Second +
				30*time.Second // buffer for overhead

			By(fmt.Sprintf("waiting for the pod to be terminated due to high CPU usage (timeout: %s)", terminationTimeout))
			GinkgoWriter.Printf("Termination timeout breakdown:\n")
			GinkgoWriter.Printf("  - Metrics collection: %s (%ds polling × %d datapoints)\n",
				metricsCollectionTime, pollingIntervalSeconds, podMetricsHistory)
			GinkgoWriter.Printf("  - Recycle delay: %ds\n", recycleDelaySeconds)
			GinkgoWriter.Printf("  - Grace period: %ds\n", gracePeriodSeconds)
			GinkgoWriter.Printf("  - Overhead buffer: 30s\n")
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

			By("verifying PodTerminated event was recorded on Recycler CR")
			verifyTerminationEvent := func() error {
				cmd = exec.Command("kubectl", "get", "events",
					"-n", testNamespace,
					"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Recycler", recyclerName),
					"-o", "jsonpath={.items[*].reason}",
				)
				events, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				eventsList := string(events)
				if !utils.ContainsString(eventsList, "PodTerminated") {
					return fmt.Errorf("PodTerminated event not found, got events: %s", eventsList)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyTerminationEvent, 30*time.Second, 2*time.Second).Should(Succeed())

			By("verifying new pod was created by deployment")
			verifyNewPodCreated := func() error {
				cmd = exec.Command("kubectl", "get", "pods",
					"-n", testNamespace,
					"-l", "app=cpu-stress",
					"-o", "jsonpath={.items[0].metadata.name}",
				)
				newPodName, err := utils.Run(cmd)
				if err != nil {
					return err
				}

				if string(newPodName) == initialPodName {
					return fmt.Errorf("pod name hasn't changed yet")
				}

				// Verify new pod is running
				cmd = exec.Command("kubectl", "get", "pod",
					string(newPodName),
					"-n", testNamespace,
					"-o", "jsonpath={.status.phase}",
				)
				status, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				if string(status) != "Running" {
					return fmt.Errorf("new pod not running yet, status: %s", status)
				}
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
