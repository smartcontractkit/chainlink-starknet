# Build image

FROM golang:1.21.5 as build

# Copy source

RUN mkdir -p /build/starknet-monitoring/cmd
COPY ./monitoring/cmd/monitoring /build/starknet-monitoring/cmd/monitoring
COPY ./monitoring/pkg /build/starknet-monitoring/pkg
COPY ./monitoring/go.mod /build/starknet-monitoring/
COPY ./monitoring/go.sum /build/starknet-monitoring/

# Copy relayer
COPY ./relayer /build/relayer

# Compile binary

WORKDIR /build/starknet-monitoring
RUN go build -o ./monitoring ./cmd/monitoring/*.go

# Production image

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /build/starknet-monitoring/monitoring /monitoring

# Expose prometheus default port
EXPOSE 9090/tcp

ENTRYPOINT ["/monitoring"]
CMD ["--help"]
