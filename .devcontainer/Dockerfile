FROM mcr.microsoft.com/devcontainers/go:2.2-1.25-bookworm@sha256:eb89cbc33cfc789d3eea538b36b8bdeda82070231bc48dd974c23e863511157e

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
