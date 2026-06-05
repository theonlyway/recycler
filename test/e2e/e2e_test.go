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
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
		It("should run successfully", operatorStartupTest)
		It("should terminate pod when CPU threshold is exceeded", cpuRecyclingTest)
	})
})
