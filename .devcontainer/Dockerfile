FROM mcr.microsoft.com/devcontainers/go:2.3-1.25-bookworm@sha256:2300b759440aa1b3dade1012ebb8cfa3baaa63bd6940e86e399c523a26f6e3cf

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
