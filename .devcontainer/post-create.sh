#!/bin/bash
set -e

echo "Installing additional Go tools..."

# Install kubebuilder
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/

# Install controller-gen
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

# Install kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
sudo mv kustomize /usr/local/bin/

# Install kind (Kubernetes in Docker)
go install sigs.k8s.io/kind@latest

# Install additional useful tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Setup bash completion for kubectl
echo 'source <(kubectl completion bash)' >> ~/.bashrc
echo 'alias k=kubectl' >> ~/.bashrc
echo 'complete -o default -F __start_kubectl k' >> ~/.bashrc

# Install envtest binaries for testing
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
setup-envtest use --use-env -p path

# Setup kind cluster
echo "Creating kind cluster..."
kind create cluster --name devcontainer --wait 5m || echo "Kind cluster may already exist"

# Configure kubectl context
echo "Configuring kubectl..."
kubectl cluster-info --context kind-devcontainer

# Verify the cluster is working
kubectl get nodes

# Set the default context
kubectl config use-context kind-devcontainer

echo "Dev container setup complete!"
echo "Kind cluster 'devcontainer' is ready!"
echo "Run 'kubectl get nodes' to verify the cluster is running."
