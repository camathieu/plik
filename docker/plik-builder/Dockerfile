FROM golang:1.13.6-buster

ARG DEBIAN_FRONTEND=noninteractive

# Install Node/NPM to build the frontend

SHELL ["/bin/bash", "-l", "-euxo", "pipefail", "-c"]

RUN apt-get update; \
    apt-get full-upgrade -y; \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
    ; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

ENV NVM_DIR /usr/local/nvm

RUN mkdir -p "$NVM_DIR"; \
    curl -o- \
        "https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh" | \
        bash \
    ; \
    source $NVM_DIR/nvm.sh; \
    nvm install --lts --latest-npm

RUN command -v nvm; \
    command -v node; \
    node --version; \
    command -v npm; \
    npm --version

# Expose the plik port
EXPOSE 8080