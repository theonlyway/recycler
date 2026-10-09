FROM mcr.microsoft.com/devcontainers/go:2.3-1.25-bookworm@sha256:4b2afc34c02ad032788a823eb38d246680249adc223ddfa273affe68e9c803eb

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
