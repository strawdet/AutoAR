# syntax=docker/dockerfile:1.7

# --- Builder stage: install Go-based security tools and build AutoAR bot ---
FROM golang:1.24-bullseye AS builder

WORKDIR /app

# Install system packages required for building tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl build-essential cmake libpcap-dev ca-certificates \
    pkg-config libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install external Go-based CLI tools used by AutoAR (nuclei, trufflehog)
# Install Nuclei
RUN GOBIN=/go/bin go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest

# Install TruffleHog
RUN git clone --depth 1 https://github.com/trufflesecurity/trufflehog.git /tmp/trufflehog && \
    cd /tmp/trufflehog && go build -o /go/bin/trufflehog . && \
    rm -rf /tmp/trufflehog

# Install Goop
RUN go install github.com/deletescape/goop@latest
# Build AutoAR main CLI and entrypoint
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies (module graph only)
RUN go mod download

# Copy application source
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build main autoar binary from cmd/autoar (CGO enabled for naabu/libpcap)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/autoar ./cmd/autoar

# Build entrypoint binary (replaces docker-entrypoint.sh)
WORKDIR /app/internal/modules/entrypoint
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/autoar-entrypoint .
WORKDIR /app

# --- Runtime stage: minimal Debian image ---
FROM debian:bullseye-slim

# NOTE: Changed AUTOAR_RESULTS_DIR to store results under a timestamped subdir
# by default; adjust AUTOAR_RESULTS_DIR at runtime if needed.
# Personal note: mounting /app/results as a volume is handy for keeping output
# across container restarts without rebuilding.
# Personal note: set AUTOAR_LOG_LEVEL=debug in your .env when troubleshooting
# scan failures - saved me a lot of head-scratching.
ENV AUTOAR_SCRIPT_PATH=/usr/local/bin/autoar \
    AUTOAR_CONFIG_FILE=/app/autoar.yaml \
    AUTOAR_RESULTS_DIR=/app/results \
    AUTOAR_LOG_LEVEL=info

WORKDIR /app

# System deps for runtime and common tools (including Java + unzip for jadx and apktool)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl ca-certificates tini jq dnsutils libpcap0.8 \
    postgresql-client docker.io \
    openjdk-17-jre-headless unzip \
    && rm -rf /var/lib/apt/lists/*

# Install jadx decompiler for apkX analysis
# Pinned to 1.5.0 (upgrade from 1.4.7) for better Kotlin decompilation support
RUN set -eux; \
    JADX_VERSION="1.5.0"; \
    curl -L "https://github.com/skylot/jadx/releases/download/v${JADX_VERSION}/jadx-${JADX_VERSION}.zip" -o /tmp/jadx.zip; \
    mkdir -p /opt/jadx; \
    unzip -q /tmp/jadx.zip -d /opt/jadx; \
    ln -sf /opt/jadx/bin/jadx /usr/local/bin/jadx; \
    ln -sf /opt/jadx/bin/jadx-gui /usr/local/bin/jadx-gui || true; \
    rm /tmp/jadx.zip

# Install apktool for MITM patching (decode/encode APKs)
# Bumped to 2.10.0 - fixes some manifest parsing issues I ran into with newer APKs
RUN set -eux; \
    APKTOOL_VERSION="2.10.0"; \
    curl -L "https://bitbucket.org/iBotPea