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
  name: name-of-recycler
  namespace: namespace-of-recycler
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: name-of-deployment
  pollingIntervalSeconds: 30
  podMetricsHistory: 5
  averageCpuUtilizationPercent: 80
  recycleDelaySeconds: 3600
  gracePeriodSeconds: 60
  metricStorageLocation: memory
  metricsSource: kubernetes
  metricsRetentionSeconds: 300
```

| Field | Default | Description |
|-------|---------|-------------|
| `scaleTargetRef.apiVersion` | `apps/v1` | API version of the target resource. |
| `scaleTargetRef.kind` | `Deployment` | Kind of the target resource. Only `Deployment` is supported. |
| `scaleTargetRef.name` | — | Name of the target Deployment to monitor. Must exist in the same namespace as the Recycler CR. |
| `averageCpuUtilizationPercent` | — | Rolling-average CPU utilization threshold as a percentage of the pod's CPU limit. Pods whose rolling average exceeds this value are marked for recycling. |
| `pollingIntervalSeconds` | `60` | How frequently (in seconds) the controller polls for CPU metrics. Lower values are more responsive but increase API server load. |
| `podMetricsHistory` | `10` | Number of polling samples kept in the rolling window used to compute average CPU utilization. Larger values smooth out short spikes; smaller values react more quickly to sustained high usage. |
| `recycleDelaySeconds` | `300` | Seconds to wait after a breach is first detected before the pod is deleted. Allows transient spikes time to recover before a recycle is triggered. |
| `gracePeriodSeconds` | `30` | Pod termination grace period in seconds. The kubelet sends SIGTERM and waits this long before sending SIGKILL. |
| `metricStorageLocation` | `memory` | Where per-pod CPU history is stored between reconcile cycles. `memory`: fast, zero API cost, lost on controller restart. `annotation`: persisted as a pod annotation, survives restarts, incurs an etcd write per poll. Only applies when `metricsSource: kubernetes`. |
| `metricsSource` | `kubernetes` | Source for per-pod CPU utilization. `kubernetes` polls the Kubernetes Metrics API; `prometheus` queries an external Prometheus server (see [Using Prometheus](#using-prometheus)) and ignores `metricStorageLocation`. |
| `metricsRetentionSeconds` | `300` | How long per-pod gauge series are retained on the `/metrics` endpoint after pod termination, allowing at least one Prometheus scrape to capture the final value. Set to `0` to remove series immediately. |
| `prometheus.serverAddress` | — | Base URL of the Prometheus server, e.g. `http://prometheus-operated.monitoring.svc:9090`. Required when `metricsSource: prometheus`. |
| `prometheus.query` | default cAdvisor query | PromQL query used to evaluate per-pod CPU utilization. Must return an instant vector with a `pod` label and a CPU percentage value. Supports Go `text/template` variables — see [Using Prometheus](#using-prometheus). When omitted, a built-in cAdvisor-only query is used. |
| `prometheus.insecureSkipVerify` | `false` | Disable TLS certificate verification when `serverAddress` uses HTTPS. |

## Metrics source

The Recycler can determine per-pod CPU utilisation from one of two sources, selected with `spec.metricsSource`:

| `metricsSource` | How it works | Requirements |
|-----------------|--------------|--------------|
| `kubernetes` (default) | Polls the Kubernetes Metrics API every `pollingIntervalSeconds`, stores a rolling history of `podMetricsHistory` samples per pod, and compares the in-process rolling average against `averageCpuUtilizationPercent`. | `metrics-server` (or equivalent Metrics API provider) installed. |
| `prometheus` | Queries an external Prometheus server each reconcile. The averaging is done by the PromQL query itself, so no per-pod history is stored by the controller (`metricStorageLocation` is ignored). | A reachable Prometheus already scraping pod CPU metrics. The default query needs only cAdvisor (always exposed by the kubelet); kube-state-metrics is not required. |

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
    query: |-
      100 * sum by (pod) (
        rate(container_cpu_usage_seconds_total{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}[{{.WindowSeconds}}s])
      ) / sum by (pod) (
        container_spec_cpu_quota{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}
        /
        container_spec_cpu_period{namespace="{{.Namespace}}", pod=~"{{.PodRegex}}", container!="", container!="POD"}
      )
```

The `query` field is rendered as a [Go `text/template`](https://pkg.go.dev/text/template) before being sent to Prometheus. Use `{{.FieldName}}` syntax to interpolate any of the following fields:

| Template variable | Example value | Description |
|-------------------|---------------|-------------|
| `{{.Namespace}}` | `my-namespace` | Namespace of the target Deployment. |
| `{{.Deployment}}` | `my-app` | Name of the target Deployment. |
| `{{.PodRegex}}` | `my-app-6b4d9f-abc\|my-app-6b4d9f-xyz` | Regex alternation of the current live pod names, with each name regex-escaped. Used with Prometheus's `=~` label matcher (`pod=~"{{.PodRegex}}"`) to scope the query to exactly the pods the controller is currently monitoring. The list is rebuilt every reconcile. |
| `{{.WindowSeconds}}` | `150` | Averaging window in seconds, computed as `podMetricsHistory × pollingIntervalSeconds`. Use this as the range selector in `rate()`/`increase()` expressions so the Prometheus window matches the rolling window the controller would use with the Kubernetes Metrics API. |

Notes:
- `spec.prometheus` is **required** when `metricsSource: prometheus` (enforced by a CEL validation on the CRD) and ignored otherwise.
- The default query expresses CPU usage as a percentage of each pod's CPU **limit**, derived purely from cAdvisor's cgroup series (`container_spec_cpu_quota / container_spec_cpu_period` == limit in cores). This deliberately avoids assuming kube-state-metrics is installed. Pods without a CPU limit have no quota series and are skipped — supply a custom `query` (e.g. one using `kube_pod_container_resource_limits`) if you measure against requests or a different source.
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

### Prerequisites

- Go v1.26.3+
- Docker (with BuildKit enabled)
- `kubectl` v1.36.0+
- [Kind](https://kind.sigs.k8s.io/) (for e2e tests)

### Development workflow

1. Fork and clone the repository.
2. Make your changes.
3. If you modified `api/v1alpha1/` types or kubebuilder marker annotations, regenerate CRDs and deepcopy methods:
   ```sh
   make generate manifests
   ```
4. Build and lint — both must pass before submitting:
   ```sh
   make build
   make lint
   ```
5. Run unit tests:
   ```sh
   make test
   ```
6. Run e2e tests (creates a temporary Kind cluster, runs the full suite, then tears it down):
   ```sh
   make test-e2e
   ```
   To manage the Kind cluster separately:
   ```sh
   make setup-test-e2e   # create the cluster
   make cleanup-test-e2e # delete the cluster
   ```

### CI checks on pull requests

All of the following checks run automatically on every pull request and must pass:

| Workflow | What it does |
|----------|-------------|
| **Lint** | Runs `go fmt` (auto-commits formatting fixes) and `golangci-lint`. |
| **Tests** | Runs `make test` (unit + integration tests using envtest) and uploads an HTML coverage report as an artifact. |
| **E2E Tests** | Spins up a Kind cluster and runs `make test-e2e`. |
| **Manifests** | Runs `make manifests` and fails if the output differs from what was committed — ensures CRDs are always in sync with the API types. |
| **Helm Chart CI** | Runs on changes to `helm-charts/`. Lints and templates the chart with Helm, then validates the output with Kubeconform against Kubernetes 1.31 schemas. Also installs the chart on a Kind cluster. |

The **Release** workflow runs on merge to `main` and handles automatic semver tagging, multi-arch image builds (linux/amd64, linux/arm64), pushing to GHCR and Docker Hub, Helm chart publishing, and generating signed SBOM attestations.

### Helm chart

The Helm chart at `helm-charts/recycler/` is **hand-maintained**. Do not run `make helm` as part of normal development — it overwrites the chart with helmify output and will clobber hand-crafted templates. After any structural change to `config/` (new resource type, RBAC change, new deployment arg), manually mirror the equivalent change in `helm-charts/recycler/templates/` and `helm-charts/recycler/values.yaml`.

Run `make help` for a full list of available targets.

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
