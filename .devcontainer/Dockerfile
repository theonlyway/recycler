FROM mcr.microsoft.com/devcontainers/go:2.0-1.25-bookworm@sha256:65af278fcedeef06eb6b672ad4a95c95cf0849cb82d4852ff014719fe6cdf262

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
