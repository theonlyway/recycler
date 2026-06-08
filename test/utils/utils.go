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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
)

const (
	prometheusOperatorVersion = "v0.86.0"
	prometheusOperatorURL     = "https://github.com/prometheus-operator/prometheus-operator/" +
		"releases/download/%s/bundle.yaml"

	certmanagerVersion = "v1.19.1"
	certmanagerURLTmpl = "https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml"

	metricsServerVersion = "v0.8.0"
	metricsServerURLTmpl = "https://github.com/kubernetes-sigs/metrics-server/releases/download/%s/components.yaml"
)

func warnError(err error) {
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "warning: %v\n", err)
}

// InstallPrometheusOperator installs the prometheus Operator to be used to export the enabled metrics.
func InstallPrometheusOperator() error {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "create", "-f", url)
	_, err := Run(cmd)
	return err
}

// Run executes the provided command within this context
func Run(cmd *exec.Cmd) ([]byte, error) {
	dir, _ := GetProjectDir()
	cmd.Dir = dir

	if err := os.Chdir(cmd.Dir); err != nil {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "chdir dir: %s\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}

	return output, nil
}

// UninstallPrometheusOperator uninstalls the prometheus
func UninstallPrometheusOperator() {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// UninstallCertManager uninstalls the cert manager
func UninstallCertManager() {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// InstallCertManager installs the cert manager bundle.
func InstallCertManager() error {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "apply", "-f", url)
	if _, err := Run(cmd); err != nil {
		return err
	}
	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
	// was re-installed after uninstalling on a cluster.
	cmd = exec.Command("kubectl", "wait", "deployment.apps/cert-manager-webhook",
		"--for", "condition=Available",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)

	_, err := Run(cmd)
	return err
}

// InstallMetricsServer installs the metrics-server for CPU/memory metrics.
func InstallMetricsServer() error {
	url := fmt.Sprintf(metricsServerURLTmpl, metricsServerVersion)
	cmd := exec.Command("kubectl", "apply", "-f", url)
	if _, err := Run(cmd); err != nil {
		return err
	}

	// Patch metrics-server to work in Kind (disable TLS verification and set faster scrape interval)
	patchJSON := `[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"},` +
		`{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-request-timeout=5s"},` +
		`{"op":"replace","path":"/spec/template/spec/containers/0/args/4","value":"--metric-resolution=10s"}]`
	cmd = exec.Command("kubectl", "patch", "deployment", "metrics-server",
		"-n", "kube-system",
		"--type=json",
		"-p", patchJSON,
	)
	if _, err := Run(cmd); err != nil {
		return err
	}

	// Debug: Check deployment status after patch
	cmd = exec.Command("kubectl", "get", "deployment", "metrics-server",
		"-n", "kube-system",
		"-o", "wide",
	)
	if output, err := Run(cmd); err == nil {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter,
			"\n=== Metrics Server Deployment Status (after patch) ===\n%s\n", string(output))
	}

	// Debug: Check metrics-server pod status before rollout
	cmd = exec.Command("kubectl", "get", "pods",
		"-n", "kube-system",
		"-l", "k8s-app=metrics-server",
		"-o", "wide",
	)
	if output, err := Run(cmd); err == nil {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "\n=== Metrics Server Pods (before rollout) ===\n%s\n", string(output))
	}

	// Wait for the rollout to complete after patching
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "\nWaiting for metrics-server rollout to complete...\n")
	cmd = exec.Command("kubectl", "rollout", "status", "deployment/metrics-server",
		"-n", "kube-system",
		"--timeout=2m",
	)
	if output, err := Run(cmd); err != nil {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Rollout status output: %s\n", string(output))
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Rollout failed, continuing to check pod status...\n")
		// Don't return error yet, check pod status first
	} else {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "Rollout completed: %s\n", string(output))
	}

	// Debug: Check metrics-server pod status after rollout
	cmd = exec.Command("kubectl", "get", "pods",
		"-n", "kube-system",
		"-l", "k8s-app=metrics-server",
		"-o", "wide",
	)
	if output, err := Run(cmd); err == nil {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "\n=== Metrics Server Pods (after rollout) ===\n%s\n", string(output))
	}

	// Debug: Get metrics-server logs
	cmd = exec.Command("kubectl", "logs",
		"-n", "kube-system",
		"deployment/metrics-server",
		"--tail=50",
	)
	if output, err := Run(cmd); err == nil {
		_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "\n=== Metrics Server Logs ===\n%s\n", string(output))
	}

	// Wait for metrics-server to be ready
	cmd = exec.Command("kubectl", "wait", "deployment/metrics-server",
		"--for", "condition=Available",
		"--namespace", "kube-system",
		"--timeout", "5m",
	)

	_, err := Run(cmd)
	return err
}

// UninstallMetricsServer uninstalls the metrics-server
func UninstallMetricsServer() {
	url := fmt.Sprintf(metricsServerURLTmpl, metricsServerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// LoadImageToKindClusterWithName loads a local docker image to the kind cluster
func LoadImageToKindClusterWithName(name string) error {
	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", name, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := Run(cmd)
	return err
}

// GetNonEmptyLines converts given command output string into individual objects
// according to line breakers, and ignores the empty elements in it.
func GetNonEmptyLines(output string) []string {
	var res []string
	elements := strings.SplitSeq(output, "\n")
	for element := range elements {
		if element != "" {
			res = append(res, element)
		}
	}

	return res
}

// GetProjectDir will return the directory where the project is
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}
	wd = strings.ReplaceAll(wd, "/test/e2e", "")
	return wd, nil
}

// StringReader creates an io.Reader from a string
func StringReader(s string) io.Reader {
	return strings.NewReader(s)
}

// ContainsString checks if a string contains a substring
func ContainsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// SetupMetricsAccess creates the ClusterRoleBinding and a short-lived service account token
// needed to access the controller's /metrics endpoint. Call this once before a sequence of
// metric fetches and reuse the returned token with FetchControllerMetricsWithToken. The
// returned cleanup function removes the ClusterRoleBinding and should be deferred by the caller.
func SetupMetricsAccess(namespace string) (token string, cleanup func(), err error) {
	const bindingName = "recycler-e2e-metrics-reader"
	bindCmd := exec.Command("kubectl", "create", "clusterrolebinding", bindingName,
		"--clusterrole=recycler-metrics-reader",
		"--serviceaccount="+namespace+":recycler-controller-manager",
	)
	if out, bindErr := bindCmd.CombinedOutput(); bindErr != nil {
		// Tolerate "already exists" so re-runs don't fail.
		if !strings.Contains(string(out), "already exists") {
			return "", func() {}, fmt.Errorf("failed to create metrics reader binding: %w — %s", bindErr, string(out))
		}
	}
	cleanup = func() {
		_ = exec.Command("kubectl", "delete", "clusterrolebinding", bindingName).Run()
	}

	tokenCmd := exec.Command("kubectl", "create", "token",
		"recycler-controller-manager", "--namespace", namespace, "--duration=600s")
	tokenBytes, tokenErr := Run(tokenCmd)
	if tokenErr != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("failed to create service account token: %w", tokenErr)
	}
	return strings.TrimSpace(string(tokenBytes)), cleanup, nil
}

// FetchControllerMetricsWithToken port-forwards the controller's metrics service and returns
// the raw Prometheus text from /metrics using the provided bearer token. For repeated calls
// (e.g. inside an Eventually loop), obtain the token once via SetupMetricsAccess and pass it
// here to avoid creating a new token on every retry.
func FetchControllerMetricsWithToken(namespace, token string) (string, error) {
	// Start port-forward in the background.
	pfCmd := exec.Command("kubectl", "port-forward",
		"svc/recycler-controller-manager-metrics-service",
		"--namespace", namespace,
		"18443:8443")
	if err := pfCmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start port-forward: %w", err)
	}
	defer func() { _ = pfCmd.Process.Kill() }()

	// Give the tunnel a moment to be established.
	time.Sleep(2 * time.Second)

	// Make an HTTPS GET with the bearer token, skipping TLS verification
	// because the manager uses a self-signed certificate.
	//nolint:gosec // G402: InsecureSkipVerify is intentional in e2e test-only code.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(http.MethodGet, "https://localhost:18443/metrics", nil)
	if err != nil {
		return "", fmt.Errorf("failed to build metrics request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("metrics request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "warning: failed to close metrics response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metrics response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metrics endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// FetchControllerMetrics is a convenience wrapper for single ad-hoc metric fetches.
// For repeated calls inside an Eventually loop, prefer SetupMetricsAccess +
// FetchControllerMetricsWithToken to avoid creating a new token on every retry.
func FetchControllerMetrics(namespace string) (string, error) {
	token, cleanup, err := SetupMetricsAccess(namespace)
	if err != nil {
		return "", err
	}
	defer cleanup()
	return FetchControllerMetricsWithToken(namespace, token)
}

// MetricValue searches raw Prometheus text for the first sample whose name and labels
// all match, returning its float64 value. Labels may be a partial set.
func MetricValue(body, metricName string, labels map[string]string) (float64, bool) {
	for _, line := range strings.Split(body, "\n") {
		// Skip comments and empty lines.
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		// Line must start with the metric name.
		if !strings.HasPrefix(line, metricName+"{") && !strings.HasPrefix(line, metricName+" ") {
			continue
		}
		// Check every required label is present.
		allMatch := true
		for k, v := range labels {
			if !strings.Contains(line, k+`="`+v+`"`) {
				allMatch = false
				break
			}
		}
		if !allMatch {
			continue
		}
		// Prometheus text format: `name{labels} value [timestamp]`
		// The value is always the second whitespace-separated token.
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		return val, true
	}
	return 0, false
}

// SumMetricValues sums the values of all samples whose name and labels all match.
// Use this for counters that are split across multiple series (e.g. per-pod label),
// where you want the total across all matching series rather than a single value.
func SumMetricValues(body, metricName string, labels map[string]string) float64 {
	var total float64
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, metricName+"{") && !strings.HasPrefix(line, metricName+" ") {
			continue
		}
		allMatch := true
		for k, v := range labels {
			if !strings.Contains(line, k+`="`+v+`"`) {
				allMatch = false
				break
			}
		}
		if !allMatch {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		total += val
	}
	return total
}
