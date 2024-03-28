#!/usr/bin/env bash

set -euo pipefail

cache_path="$(git rev-parse --show-toplevel)/.local-mock-server"
binary_name="dummy-external-adapter"
binary_path="${cache_path}/bin/${binary_name}"

bash "$(dirname -- "$0")/mock-adapter.down.sh"

listen_address="0.0.0.0:6060"
echo "Listen address: ${listen_address}"

if [ ! -f "${binary_path}" ]; then
	echo "Installing mock-adapter"
	export GOPATH="${cache_path}"
	export GOBIN="${cache_path}/bin"
	go install 'github.com/smartcontractkit/dummy-external-adapter@latest'
fi

nohup "${binary_path}" "${listen_address}" &>/dev/null &
echo "Started mock-adapter (PID $!)"
