FROM mcr.microsoft.com/devcontainers/go:2.2-1.25-bookworm@sha256:3aa036e798dfa6a86f21b0de813111815f21efd21ef43ea75c85696c22e59d08

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
