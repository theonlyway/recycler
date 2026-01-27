# Kubernetes Controller Development Container

This dev container provides a complete environment for developing Kubernetes controllers in Go.

## Included Tools

- **Go** (latest 1.22.x)
- **kubectl** - Kubernetes CLI
- **kind** - Kubernetes in Docker for local testing
- **kubebuilder** - SDK for building Kubernetes APIs
- **controller-gen** - Generate CRD manifests and RBAC
- **kustomize** - Kubernetes configuration management
- **golangci-lint** - Go linter aggregator
- **Docker-in-Docker** - Build and run containers
- **Ginkgo** - BDD testing framework
- **envtest** - Test environment for controllers

## Getting Started

1. Open this repository in VS Code
2. When prompted, click "Reopen in Container"
3. Wait for the container to build and post-create script to complete
4. Start developing your controller!

## Common Commands

```bash
# Initialize a new controller project
kubebuilder init --domain example.com --repo github.com/yourorg/yourproject

# Create a new API
kubebuilder create api --group apps --version v1 --kind MyApp

# Create a local Kind cluster
kind create cluster

# Run tests
make test

# Build and deploy
make docker-build docker-push IMG=yourregistry/controller:tag
make deploy IMG=yourregistry/controller:tag
