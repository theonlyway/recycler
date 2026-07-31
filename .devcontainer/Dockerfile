FROM mcr.microsoft.com/devcontainers/go:2.1-1.25-bookworm@sha256:30cbde1de62f190904446d8b81dcfd0befcfd60889419edb8872d598be9632cc

# Fix the Yarn GPG key issue by removing the Yarn repository
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*
