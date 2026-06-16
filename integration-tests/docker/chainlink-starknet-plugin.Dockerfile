# Overlay image: chainlink-plugins base + locally built chainlink-starknet plugin.
# Build via integration-tests/scripts/build-chainlink-image.sh

ARG CHAINLINK_PLUGINS_BASE=chainlink-plugins:local-base
FROM golang:1.26.2-bookworm AS build-starknet-plugin

WORKDIR /src/relayer
COPY relayer/go.mod relayer/go.sum ./

ARG CL_GOPRIVATE=github.com/smartcontractkit/*
ENV GOPRIVATE="${CL_GOPRIVATE}"

RUN --mount=type=secret,id=GIT_AUTH_TOKEN \
    set -e && \
    if [ -f /run/secrets/GIT_AUTH_TOKEN ]; then \
      git config --global url."https://x-access-token:$(cat /run/secrets/GIT_AUTH_TOKEN)@github.com/".insteadOf "https://github.com/"; \
    fi && \
    go mod download

COPY relayer/ ./

RUN CGO_ENABLED=0 go build -ldflags="-s" -o /chainlink-starknet ./pkg/chainlink/cmd/chainlink-starknet

FROM ${CHAINLINK_PLUGINS_BASE}

COPY --from=build-starknet-plugin /chainlink-starknet /usr/local/bin/chainlink-starknet
ENV CL_STARKNET_CMD=chainlink-starknet
