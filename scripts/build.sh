#!/usr/bin/env bash

set -euo pipefail

container_version="starknet"

pushd "$(dirname -- "$0")/../core"
docker build . -t smartcontract/chainlink:${container_version} -f ./core/chainlink.Dockerfile
popd

