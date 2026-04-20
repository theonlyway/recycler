FROM mcr.microsoft.com/devcontainers/go:2.1-1.25-bookworm@sha256:80694b9e519d41822772c5e4b2f7903343f7591ca5db064613ec7e53b855dc31

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
