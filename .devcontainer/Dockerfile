FROM mcr.microsoft.com/devcontainers/go:2.1-1.25-bookworm@sha256:f8f841c474489d6cd5b67cb30ee0e56ae5f1379f0bdf2679191573c41b5e11f8

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
