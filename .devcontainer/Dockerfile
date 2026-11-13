FROM mcr.microsoft.com/devcontainers/go:2.3-1.25-bookworm@sha256:328e54b6b9ee08c44e80c51157e92c9b6ddedfe5ece35a374aaa82b91af5c3ac

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
