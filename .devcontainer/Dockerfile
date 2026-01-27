FROM mcr.microsoft.com/devcontainers/go:2.0-bookworm@sha256:ef7d7fe9f8eb237c2d6b8403bc0affceca3c349a041fb7714dfc554d9a97adca

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
