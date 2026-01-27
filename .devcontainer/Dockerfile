FROM mcr.microsoft.com/devcontainers/go:2.0-1.25-bookworm

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
