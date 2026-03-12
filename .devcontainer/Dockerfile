FROM mcr.microsoft.com/devcontainers/go:2.1-1.25-bookworm@sha256:2765adcd37710f2e24c67837cf9c88ae1d88708f405c15a88b307711268db63e

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
