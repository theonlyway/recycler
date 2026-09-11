FROM mcr.microsoft.com/devcontainers/go:2.3-1.25-bookworm@sha256:bec993b44ced2156deafe4fb05af7fc0bbe1aca88ed5d710c305654a735fb558

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
