# Build image

FROM golang:1.20-buster as build

# Copy source

RUN mkdir -p /starknet-monitoring/cmd
COPY ./cmd/monitoring /starknet-monitoring/cmd/monitoring
COPY ./pkg /starknet-monitoring/pkg
COPY ./go.mod /starknet-monitoring/
COPY ./go.sum /starknet-monitoring/

# Compile binary

WORKDIR /starknet-monitoring
RUN go build -o ./monitoring ./cmd/monitoring/*.go

# Production image

FROM ubuntu:20.04

RUN apt-get update && apt-get install -y ca-certificates
COPY --from=build /starknet-monitoring/monitoring /monitoring

# Expose prometheus default port
EXPOSE 9090/tcp

ENTRYPOINT ["/monitoring"]
CMD ["--help"]
