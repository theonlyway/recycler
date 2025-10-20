[![Build and Push Recycler Operator image](https://github.com/theonlyway/recycler/actions/workflows/build.yml/badge.svg)](https://github.com/theonlyway/recycler/actions/workflows/build.yml)
[![Lint](https://github.com/theonlyway/recycler/actions/workflows/lint.yml/badge.svg)](https://github.com/theonlyway/recycler/actions/workflows/lint.yml)
[![Tests](https://github.com/theonlyway/recycler/actions/workflows/test.yml/badge.svg)](https://github.com/theonlyway/recycler/actions/workflows/test.yml)
[![E2E Tests](https://github.com/theonlyway/recycler/actions/workflows/test-e2e.yml/badge.svg)](https://github.com/theonlyway/recycler/actions/workflows/test-e2e.yml)
# recycler

A Kubernetes controller that monitors pods CPU utilisation inside a deployment, replicaset, or statefulset and terminates the pod if it exceeds a specified threshold.

## Description
Ideally something like this shouldn't even exist if people wrote their software properly. But sometimes bugs exist for longer than they should and you get sick of a HPA scaling needlessly, a pod not failing health checks even though it's at 100% CPU, and one day you are on leave and you've hit the limit you set on the HPA. All this results in some graph being more red then it should be which causes someone to panic. Until someone fixes their bug in the code, this controller was created to monitor pods and terminate them if they exceed a defined threshold.

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

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
