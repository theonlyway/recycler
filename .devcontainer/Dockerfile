FROM mcr.microsoft.com/devcontainers/go:2.1-1.25-bookworm@sha256:50f86fc75fb76e29cfa69bd11685d4915ef58bacdb11d62921b3cc5f58a152e4

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
