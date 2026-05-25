FROM mcr.microsoft.com/devcontainers/go:2.1-1.25-bookworm@sha256:768f4d177236e543dafba0f140754ac1d084651ee70c88b3e4ff90e2034e592f

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
