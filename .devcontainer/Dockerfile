# Optional: Use this if you need additional customization beyond the base image
FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm

# Install additional system dependencies if needed
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    build-essential \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*
