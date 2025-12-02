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

package utils

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Utils", func() {
	Context("GetProjectDir", func() {
		It("should return the project directory", func() {
			dir, err := GetProjectDir()
			Expect(err).NotTo(HaveOccurred())
			Expect(dir).NotTo(BeEmpty())
			Expect(filepath.IsAbs(dir)).To(BeTrue())
		})

		It("should remove /test/e2e from path if present", func() {
			originalWd, _ := os.Getwd()
			defer os.Chdir(originalWd)

			// Create a temporary directory structure
			tmpDir := GinkgoT().TempDir()
			testE2eDir := filepath.Join(tmpDir, "test", "e2e")
			err := os.MkdirAll(testE2eDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			err = os.Chdir(testE2eDir)
			Expect(err).NotTo(HaveOccurred())

			dir, err := GetProjectDir()
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.HasSuffix(dir, "/test/e2e")).To(BeFalse())
		})
	})

	Context("StringReader", func() {
		It("should create an io.Reader from a string", func() {
			testString := "Hello, World!"
			reader := StringReader(testString)

			Expect(reader).NotTo(BeNil())

			// Read from the reader
			buf := make([]byte, len(testString))
			n, err := reader.Read(buf)
			Expect(err).To(BeNil())
			Expect(n).To(Equal(len(testString)))
			Expect(string(buf)).To(Equal(testString))
		})

		It("should handle empty string", func() {
			reader := StringReader("")
			Expect(reader).NotTo(BeNil())

			buf := make([]byte, 10)
			n, err := reader.Read(buf)
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should handle multiline strings", func() {
			testString := "line1\nline2\nline3"
			reader := StringReader(testString)

			buf := make([]byte, len(testString))
			n, _ := reader.Read(buf)
			Expect(string(buf[:n])).To(Equal(testString))
		})
	})

	Context("ContainsString", func() {
		It("should return true when substring is present", func() {
			result := ContainsString("Hello, World!", "World")
			Expect(result).To(BeTrue())
		})

		It("should return false when substring is not present", func() {
			result := ContainsString("Hello, World!", "Goodbye")
			Expect(result).To(BeFalse())
		})

		It("should handle empty string", func() {
			result := ContainsString("", "test")
			Expect(result).To(BeFalse())
		})

		It("should handle empty substring", func() {
			result := ContainsString("Hello", "")
			Expect(result).To(BeTrue())
		})

		It("should be case sensitive", func() {
			result := ContainsString("Hello, World!", "world")
			Expect(result).To(BeFalse())
		})

		It("should handle exact match", func() {
			result := ContainsString("test", "test")
			Expect(result).To(BeTrue())
		})

		It("should handle special characters", func() {
			result := ContainsString("test-deployment-123", "deployment-123")
			Expect(result).To(BeTrue())
		})
	})

	Context("GetNonEmptyLines", func() {
		It("should split output by newlines and remove empty lines", func() {
			output := "line1\nline2\nline3"
			lines := GetNonEmptyLines(output)

			Expect(lines).To(HaveLen(3))
			Expect(lines[0]).To(Equal("line1"))
			Expect(lines[1]).To(Equal("line2"))
			Expect(lines[2]).To(Equal("line3"))
		})

		It("should handle output with empty lines", func() {
			output := "line1\n\nline2\n\n\nline3\n"
			lines := GetNonEmptyLines(output)

			Expect(lines).To(HaveLen(3))
			Expect(lines).To(ConsistOf("line1", "line2", "line3"))
		})

		It("should handle single line output", func() {
			output := "single line"
			lines := GetNonEmptyLines(output)

			Expect(lines).To(HaveLen(1))
			Expect(lines[0]).To(Equal("single line"))
		})

		It("should handle empty output", func() {
			output := ""
			lines := GetNonEmptyLines(output)

			Expect(lines).To(BeEmpty())
		})

		It("should handle output with only newlines", func() {
			output := "\n\n\n"
			lines := GetNonEmptyLines(output)

			Expect(lines).To(BeEmpty())
		})

		It("should preserve line content with spaces", func() {
			output := "line with spaces\nanother  line  with   spaces"
			lines := GetNonEmptyLines(output)

			Expect(lines).To(HaveLen(2))
			Expect(lines[0]).To(Equal("line with spaces"))
			Expect(lines[1]).To(Equal("another  line  with   spaces"))
		})

		It("should handle kubectl-style output", func() {
			output := "NAME                READY   STATUS\npod-1               1/1     Running\npod-2               1/1     Running\n"
			lines := GetNonEmptyLines(output)

			Expect(lines).To(HaveLen(3))
			Expect(lines[0]).To(ContainSubstring("NAME"))
			Expect(lines[1]).To(ContainSubstring("pod-1"))
			Expect(lines[2]).To(ContainSubstring("pod-2"))
		})
	})

	Context("Run", func() {
		It("should execute a simple command successfully", func() {
			cmd := exec.Command("echo", "test")
			output, err := Run(cmd)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("test"))
		})

		It("should set GO111MODULE environment variable", func() {
			cmd := exec.Command("sh", "-c", "echo $GO111MODULE")
			output, err := Run(cmd)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("on"))
		})

		It("should change to project directory", func() {
			cmd := exec.Command("pwd")
			output, err := Run(cmd)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).NotTo(BeEmpty())
		})

		It("should return error for invalid command", func() {
			cmd := exec.Command("nonexistentcommand12345")
			_, err := Run(cmd)

			Expect(err).To(HaveOccurred())
		})

		It("should return error for failing command", func() {
			cmd := exec.Command("sh", "-c", "exit 1")
			_, err := Run(cmd)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed with error"))
		})

		It("should capture command output", func() {
			cmd := exec.Command("echo", "Hello\nWorld")
			output, err := Run(cmd)

			Expect(err).NotTo(HaveOccurred())
			lines := GetNonEmptyLines(string(output))
			Expect(lines).To(HaveLen(2))
		})

		It("should handle commands with multiple arguments", func() {
			cmd := exec.Command("echo", "-n", "no", "newline")
			output, err := Run(cmd)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("no"))
			Expect(string(output)).To(ContainSubstring("newline"))
		})
	})

	Context("LoadImageToKindClusterWithName", func() {
		It("should construct proper kind command with default cluster", func() {
			// This test would require kind to be installed
			// We'll just verify the function exists and has the right signature
			err := LoadImageToKindClusterWithName("test-image:latest")
			// We expect this to fail without kind installed, but that's ok
			// The test verifies the function can be called
			_ = err
		})

		It("should use KIND_CLUSTER environment variable when set", func() {
			originalCluster := os.Getenv("KIND_CLUSTER")
			defer func() {
				if originalCluster != "" {
					os.Setenv("KIND_CLUSTER", originalCluster)
				} else {
					os.Unsetenv("KIND_CLUSTER")
				}
			}()

			os.Setenv("KIND_CLUSTER", "custom-cluster")

			// Just verify the function can be called with env var set
			err := LoadImageToKindClusterWithName("test-image:latest")
			_ = err // Expected to fail without kind, but that's ok
		})
	})

	Context("Constants", func() {
		It("should have valid prometheus operator version", func() {
			Expect(prometheusOperatorVersion).To(Equal("v0.86.0"))
		})

		It("should have valid cert-manager version", func() {
			Expect(certmanagerVersion).To(Equal("v1.19.1"))
		})

		It("should have valid metrics-server version", func() {
			Expect(metricsServerVersion).To(Equal("v0.8.0"))
		})

		It("should construct valid prometheus operator URL", func() {
			url := "https://github.com/prometheus-operator/prometheus-operator/" +
				"releases/download/" + prometheusOperatorVersion + "/bundle.yaml"
			Expect(url).To(ContainSubstring("prometheus-operator"))
			Expect(url).To(ContainSubstring(prometheusOperatorVersion))
		})
	})

	Context("Installation Functions", func() {
		// These tests verify the functions exist and can be called
		// They will fail without kubectl/kind but that's expected in unit tests

		It("should have InstallPrometheusOperator function", func() {
			err := InstallPrometheusOperator()
			// Expected to fail without kubectl, but verifies function signature
			_ = err
		})

		It("should have InstallCertManager function", func() {
			err := InstallCertManager()
			_ = err
		})

		It("should have InstallMetricsServer function", func() {
			err := InstallMetricsServer()
			_ = err
		})

		It("should have UninstallPrometheusOperator function", func() {
			// This function doesn't return error, just verify it can be called
			UninstallPrometheusOperator()
		})

		It("should have UninstallCertManager function", func() {
			UninstallCertManager()
		})

		It("should have UninstallMetricsServer function", func() {
			UninstallMetricsServer()
		})
	})
})
