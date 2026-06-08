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
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/go-logr/logr"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	recyclertheonlywayecomv1alpha1 "github.com/theonlyway/recycler/api/v1alpha1"
)

// defaultPrometheusQuery computes per-pod CPU utilization as a percentage of the pod's CPU limit,
// averaged over the configured window. It depends only on cAdvisor metrics (exposed by every
// kubelet), so it does not assume kube-state-metrics is present: usage comes from
// container_cpu_usage_seconds_total and the limit (in cores) is derived from the cgroup
// quota/period series (container_spec_cpu_quota / container_spec_cpu_period). Pods without a CPU
// limit produce no quota series and are skipped.
const defaultPrometheusQuery = `100 * sum by (pod) (` +
	`rate(container_cpu_usage_seconds_total{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}[{{.WindowSeconds}}s])` +
	`) / sum by (pod) (` +
	`container_spec_cpu_quota{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}` +
	` / ` +
	`container_spec_cpu_period{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}` +
	`)`

// promQueryData holds the fields exposed to the PromQL query template.
type promQueryData struct {
	Namespace     string
	Deployment    string
	PodRegex      string
	WindowSeconds int64
}

// queryPrometheusCPUUtilization queries the configured Prometheus server and returns a map of
// pod name to CPU utilization percentage for the supplied pods.
func queryPrometheusCPUUtilization(
	ctx context.Context,
	recycler *recyclertheonlywayecomv1alpha1.Recycler,
	namespace string,
	deploymentName string,
	podNames []string,
	log logr.Logger,
) (map[string]float64, error) {
	promSpec := recycler.Spec.Prometheus
	if promSpec == nil {
		return nil, fmt.Errorf("prometheus configuration is nil but metricsSource is prometheus")
	}
	if len(podNames) == 0 {
		return map[string]float64{}, nil
	}

	query, err := renderPrometheusQuery(recycler, namespace, deploymentName, podNames)
	if err != nil {
		return nil, fmt.Errorf("failed to render PromQL query: %w", err)
	}

	roundTripper := promapi.DefaultRoundTripper
	if promSpec.InsecureSkipVerify {
		transport := promapi.DefaultRoundTripper.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- opt-in via spec.prometheus.insecureSkipVerify
		roundTripper = transport
	}

	apiClient, err := promapi.NewClient(promapi.Config{
		Address:      promSpec.ServerAddress,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	v1api := promv1.NewAPI(apiClient)
	log.V(1).Info("Querying Prometheus for CPU utilization", "controller", monitorControllerName, "address", promSpec.ServerAddress, "query", query)

	result, warnings, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}
	for _, w := range warnings {
		log.Info("Prometheus query warning", "controller", monitorControllerName, "warning", w)
	}

	vector, ok := result.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected Prometheus result type %s, expected an instant vector", result.Type())
	}

	utilization := make(map[string]float64, len(vector))
	for _, sample := range vector {
		podName := string(sample.Metric["pod"])
		if podName == "" {
			log.V(1).Info("Skipping Prometheus sample without a pod label", "controller", monitorControllerName, "metric", sample.Metric.String())
			continue
		}
		if sample.Value.String() == "NaN" {
			continue
		}
		utilization[podName] = float64(sample.Value)
	}

	log.V(1).Info("Fetched CPU utilization from Prometheus", "controller", monitorControllerName, "pods", len(utilization))
	return utilization, nil
}

// renderPrometheusQuery renders the configured (or default) PromQL query template.
func renderPrometheusQuery(
	recycler *recyclertheonlywayecomv1alpha1.Recycler,
	namespace string,
	deploymentName string,
	podNames []string,
) (string, error) {
	queryTemplate := recycler.Spec.Prometheus.Query
	if strings.TrimSpace(queryTemplate) == "" {
		queryTemplate = defaultPrometheusQuery
	}

	windowSeconds := int64(recycler.Spec.PodMetricsHistory) * int64(recycler.Spec.PollingIntervalSeconds)
	if windowSeconds <= 0 {
		windowSeconds = 60
	}

	data := promQueryData{
		Namespace:     namespace,
		Deployment:    deploymentName,
		PodRegex:      buildPodRegex(podNames),
		WindowSeconds: windowSeconds,
	}

	tmpl, err := template.New("promQuery").Parse(queryTemplate)
	if err != nil {
		return "", err
	}

	var rendered strings.Builder
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", err
	}
	return rendered.String(), nil
}

// buildPodRegex builds a regex alternation of the supplied pod names, escaping each name so it
// can be safely embedded in a PromQL pod=~"..." matcher.
func buildPodRegex(podNames []string) string {
	escaped := make([]string, 0, len(podNames))
	for _, name := range podNames {
		escaped = append(escaped, regexp.QuoteMeta(name))
	}
	return strings.Join(escaped, "|")
}
