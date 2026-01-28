FROM mcr.microsoft.com/devcontainers/go:2.0-1.25-bookworm@sha256:0624dcde5d7b52a2cd0ad55de4fb8ecaf2dd32f09e564c987871dece943c7690

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
