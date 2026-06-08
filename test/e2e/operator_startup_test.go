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
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/theonlyway/recycler/test/utils"
)

func operatorStartupTest() {
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
}
