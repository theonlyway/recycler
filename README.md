[![Build and Push Recycler Operator image](https://github.com/theonlyway/recycler/actions/workflows/build.yml/badge.svg)](https://github.com/theonlyway/recycler/actions/workflows/build.yml)
# recycler

A Kubernetes controller that monitors pods CPU utilisation inside a deployment, replicaset, or statefulset and terminates the pod if it exceeds a specified threshold.

## Description
Ideally something like this shouldn't even exist if people wrote their software properly. But sometimes bugs exist for longer than they should and you get sick of a HPA scaling needlessly, a pod not failing health checks even though it's at 100% CPU, and one day you are on leave and you've hit the limit you set on the HPA. All this results in some graph being more red then it should be which causes someone to panic. Until someone fixes their bug in the code, this controller was created to monitor pods and terminate them if they exceed a defined threshold.

### Prerequisites
- go version v1.26.3+
- docker version 17.03+.
- kubectl version v1.36.0+.
- Access to a Kubernetes v1.36.0+ cluster.

## Automatic installation
### Helm
**Install the operator from `ghcr`:**
```sh
helm install recycler oci://ghcr.io/theonlyway/charts/recycler --namespace <namespace> --create-namespace
```

**Download a copy of the chart files locally from `ghcr`:**
```sh
helm pull oci://ghcr.io/theonlyway/charts/recycler --version <version>
```

**Install a specific version of the operator from `ghcr`:**
```sh
helm install recycler oci://ghcr.io/theonlyway/charts/recycler --namespace <namespace> --create-namespace --version <version>
```

**Upgrade the operator from `ghcr`:**
```sh
helm upgrade recycler oci://ghcr.io/theonlyway/charts/recycler --namespace <namespace>
```

**Upgrade to a specific version of the operator from `ghcr`:**
```sh
helm upgrade recycler oci://ghcr.io/theonlyway/charts/recycler --namespace <namespace> --version <version>
```

**Uninstall the operator:**
```sh
helm uninstall recycler --namespace <namespace>
```

### Kustomize
**Clone the repository:**
```sh
git clone https://github.com/theonlyway/recycler.git
cd recycler
```

**Install the CRDs into the cluster:**
```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**
```sh
make deploy IMG=ghcr.io/theonlyway/recycler:latest
```

**Uninstall the CRDs from the cluster:**
```sh
make uninstall
```

**UnDeploy the controller from the cluster:**
```sh
make undeploy
```

**Generate a consolidated YAML with CRDs and deployment:**
```sh
make build-installer IMG=ghcr.io/theonlyway/recycler:latest
```

The generated YAML file will be located in the `dist/install.yaml` file. You can apply it to your cluster using:
```sh
kubectl apply -f dist/install.yaml
```

## Custom Resource Definition
These are the configurable values for the Recycler custom resource. View the openAPI schema [here](config/crd/bases/recycler.theonlywaye.com_recyclers.yaml).
```yaml
apiVersion: recycler.theonlywaye.com/v1alpha1
kind: Recycler
metadata:
  name: name-of-recycler # Should be unique but can be anything you want
  namespace: namespace-of-recycler # Should be the same as the namespace of the deployment, replicaset, or statefulset
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: name-of-deployment # Should be the same as the name of the deployment, replicaset, or statefulset
  pollingIntervalSeconds: 30 # This is how long between polling for metrics from the metrics api
  podMetricsHistory: 5 # This is how many historical metrics to keep which is used to calculate the average CPU averageCpuUtilizationPercent
  averageCpuUtilizationPercent: 80 # This is the threshold for when to terminate the pod
  recycleDelaySeconds: 3600 # This is how long to wait before terminating the pod once it's breached the average CPU utilization threshold
  gracePeriodSeconds: 60 # Configuraable time to wait when terminating the pod before it's forcefully terminated
  metricStorageLocation: memory # Where to store the metrics data. Either in memory or as an annotation on the pod. There are implications to both
  metricsSource: kubernetes # Where CPU utilisation comes from: "kubernetes" (default, Metrics API) or "prometheus"
```

## Metrics source

The Recycler can determine per-pod CPU utilisation from one of two sources, selected with `spec.metricsSource`:

| `metricsSource` | How it works | Requirements |
|-----------------|--------------|--------------|
| `kubernetes` (default) | Polls the Kubernetes Metrics API every `pollingIntervalSeconds`, stores a rolling history of `podMetricsHistory` samples per pod, and compares the in-process rolling average against `averageCpuUtilizationPercent`. | `metrics-server` (or equivalent Metrics API provider) installed. |
| `prometheus` | Queries an external Prometheus server each reconcile. The averaging is done by the PromQL query itself, so no per-pod history is stored by the controller (`metricStorageLocation` is ignored). | A reachable Prometheus already scraping pod CPU metrics (e.g. cAdvisor + kube-state-metrics). |

> Both sources produce identical behaviour, events, and `/metrics` output — only the measurement source differs. The `recycleDelaySeconds` delay, breach/recovery detection, and pod termination logic are the same.

### Using Prometheus

Set `metricsSource: prometheus` and provide a `prometheus` block. This is useful when Prometheus is already monitoring CPU and you'd rather reuse it than poll the Metrics API.

```yaml
apiVersion: recycler.theonlywaye.com/v1alpha1
kind: Recycler
metadata:
  name: name-of-recycler
  namespace: namespace-of-recycler
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: name-of-deployment
  pollingIntervalSeconds: 30
  podMetricsHistory: 5 # Used only to size the averaging window: podMetricsHistory * pollingIntervalSeconds
  averageCpuUtilizationPercent: 80
  recycleDelaySeconds: 3600
  gracePeriodSeconds: 60
  metricsSource: prometheus
  prometheus:
    serverAddress: http://prometheus-operated.monitoring.svc:9090 # Base URL of the Prometheus server (required)
    insecureSkipVerify: false # Optional: disable TLS verification for an HTTPS serverAddress
    # query is optional. When omitted, a default cAdvisor + kube-state-metrics query is used.
    # It must return an instant vector where each sample has a "pod" label and a CPU
    # utilisation percentage value. The query is rendered as a Go text/template with:
    #   {{.Namespace}}      - namespace of the target Deployment
    #   {{.Deployment}}     - name of the target Deployment
    #   {{.PodRegex}}       - regex alternation of the current pod names (pod1|pod2|...)
    #   {{.WindowSeconds}}  - podMetricsHistory * pollingIntervalSeconds, the averaging window
    query: |-
      100 * sum by (pod) (
        rate(container_cpu_usage_seconds_total{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}[{{.WindowSeconds}}s])
      ) / sum by (pod) (
        kube_pod_container_resource_limits{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", resource="cpu"}
      )
```

Notes:
- `spec.prometheus` is **required** when `metricsSource: prometheus` (enforced by a CEL validation on the CRD) and ignored otherwise.
- The default query expresses CPU usage as a percentage of each pod's CPU **limit** (via `kube_pod_container_resource_limits`). Pods without a CPU limit return `NaN` and are skipped — supply a custom `query` if you measure against requests or use a different metric source.
- Each reconcile issues a **single** Prometheus query containing all current pod names, not one query per pod.


## Prometheus Metrics

The controller exposes the following custom metrics on the `/metrics` endpoint (HTTPS, port `8443`). If you are using the Prometheus Operator, set the Helm value `prometheus.serviceMonitor.enabled=true` to deploy a `ServiceMonitor` and enable scraping.

The `ServiceMonitor` must carry the labels that your Prometheus instance selects on. Check your Prometheus CR's `serviceMonitorSelector` to determine the required labels:
```sh
kubectl get prometheus -A -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}: {.spec.serviceMonitorSelector}{"\n"}{end}'
```

Then pass the required labels via `prometheus.serviceMonitor.additionalLabels`. For example, if the output is `{"matchLabels":{"release":"kube-prometheus-stack"}}`:
```sh
helm install recycler oci://ghcr.io/theonlyway/charts/recycler \
  --namespace recycler-system --create-namespace \
  --set prometheus.serviceMonitor.enabled=true \
  --set prometheus.serviceMonitor.additionalLabels.release=kube-prometheus-stack
```

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `recycler_pod_recycles_total` | Counter | `recycler_namespace`, `recycler` | Total number of pods deleted by the controller. Increments each time a pod is terminated after breaching the CPU threshold. |
| `recycler_cpu_threshold_breaches_total` | Counter | `recycler_namespace`, `recycler` | Total number of CPU threshold breach events detected. Increments when a pod first crosses the threshold and the breach annotation is written. |
| `recycler_cpu_breach_duration_seconds` | Histogram | `recycler_namespace`, `recycler` | Time in seconds between when the breach annotation was written and when the pod was actually deleted (i.e. how long the pod spent above threshold before recycling). Buckets: `30, 60, 120, 180, 300, 600, 900, 1800`. |
| `recycler_pod_last_recycle_timestamp_seconds` | Gauge | `recycler_namespace`, `recycler`, `recycler_pod` | Unix timestamp of the most recent recycle event for a specific pod. Useful for building an audit history of which pods were terminated and when. |
| `recycler_pod_cpu_utilization_percent` | Gauge | `recycler_namespace`, `recycler_pod` | Current CPU utilisation percentage for each monitored pod. For `metricsSource: kubernetes` this is the rolling average over the `podMetricsHistory` window; for `metricsSource: prometheus` it is the value returned by the configured PromQL query. |

### Example queries

**Rate of pod recycles per namespace:**
```promql
rate(recycler_pod_recycles_total[5m])
```

**Pods currently above threshold (utilisation gauge):**
```promql
recycler_pod_cpu_utilization_percent > <threshold>
```

**95th percentile breach-to-recycle duration:**
```promql
histogram_quantile(0.95, rate(recycler_cpu_breach_duration_seconds_bucket[1h]))
```

**Total breaches detected by recycler CR:**
```promql
recycler_cpu_threshold_breaches_total
```

## Building and deploying manually
### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=ghcr.io/theonlyway/recycler:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=ghcr.io/theonlyway/recycler:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/recycler:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/recycler/<tag or branch>/dist/install.yaml
```

## Security & Verification

Each release includes cryptographically signed build provenance attestations and SBOMs (Software Bill of Materials) in both SPDX and CycloneDX formats. These are attached to each GitHub release and pushed to the container registries.

### Verify Attestations

Requires the [GitHub CLI](https://cli.github.com/).

**Verify build provenance (GHCR):**
```sh
gh attestation verify oci://ghcr.io/theonlyway/recycler:<version> \
  --repo theonlyway/recycler
```

**Verify SBOM attestation (GHCR):**
```sh
gh attestation verify oci://ghcr.io/theonlyway/recycler:<version> \
  --repo theonlyway/recycler \
  --predicate-type https://spdx.dev/Document/v2.3
```

**View full attestation details:**
```sh
gh attestation verify oci://ghcr.io/theonlyway/recycler:<version> \
  --repo theonlyway/recycler \
  --format json | jq
```

### SBOM Files

Four SBOM files are attached to each release:

| File | Format | Image |
|------|--------|-------|
| `sbom-ghcr.spdx.json` | SPDX 2.3 | GHCR |
| `sbom-ghcr.cyclonedx.json` | CycloneDX | GHCR |
| `sbom-dockerhub.spdx.json` | SPDX 2.3 | Docker Hub |
| `sbom-dockerhub.cyclonedx.json` | CycloneDX | Docker Hub |

Use **SPDX** for attestation verification and compliance. Use **CycloneDX** with security scanning tools like [Grype](https://github.com/anchore/grype), [Trivy](https://github.com/aquasecurity/trivy), or [Dependency-Track](https://dependencytrack.org/).

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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
