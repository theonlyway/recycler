FROM mcr.microsoft.com/devcontainers/go:2.3-1.25-bookworm@sha256:25807d3f171f858cfedc349651a16de7387cb051c9a779b8ced17c45e744eb62

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
