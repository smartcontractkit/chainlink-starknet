#!/usr/bin/env bash
# TODO: this script needs to be replaced with a predefined K8s enviroment

set -euo pipefail

# cpu_struct=`arch`;
# echo $cpu_struct;
cpu_struct="linux"

# Clean up first
bash "$(dirname -- "$0")/devnet-hardhat-down.sh"

echo "Checking CPU structure..."
if [[ $cpu_struct == *"arm"* ]]; then
	echo "Starting arm devnet container..."
	container_version="${CONTAINER_VERSION:-a147b4cd72f9ce9d1fa665d871231370db0f51c7}-arm"
else
	echo "Starting i386 devnet container..."
	container_version="${CONTAINER_VERSION:-a147b4cd72f9ce9d1fa665d871231370db0f51c7}"
fi

echo "Starting starknet-devnet"

# we need to replace the entrypoint because starknet-devnet's docker builds at 0.5.1 don't include cargo or gcc.
docker run \
	-p 127.0.0.1:5050:5050 \
	-p 127.0.0.1:8545:8545 \
	-d \
	-e RUST_LOG=debug \
	--name chainlink-starknet.starknet-devnet \
	"shardlabs/starknet-devnet-rs:${container_version}" \
	--seed 0 \
	--gas-price 1 \
	--data-gas-price 1 \
	--account-class cairo1

echo "Starting hardhat..."
docker run --net container:chainlink-starknet.starknet-devnet -d --name chainlink-starknet.hardhat ethereumoptimism/hardhat-node:nightly

wait_for_container() {
	local container_name="$1"
	local ready_log="$2"
	local start_time=$(date +%s)
	local prev_output=""

	echo "Waiting for container $container_name to become ready.."
	while true; do
		output=$(docker logs "$container_name" 2>&1)
		if [[ "${output}" != "${prev_output}" ]]; then
			echo -n "${output#$prev_output}"
			prev_output="${output}"
		fi

		if [[ $output == *"$ready_log"* ]]; then
			echo ""
			echo "container $container_name is ready."
			return
		fi

		current_time=$(date +%s)
		elapsed_time=$((current_time - start_time))
		if ((elapsed_time > 600)); then
			echo "Error: Command did not become ready within 600 seconds"
			exit 1
		fi

		sleep 3
	done
}

# starknet-devnet startup is slow and requires compiling cairo.
wait_for_container "chainlink-starknet.starknet-devnet" "listening"

# ethereumoptimism/hardhat-node is also slow and should be online before l1-l2 messaging tests are run
wait_for_container "chainlink-starknet.hardhat" "Any funds sent to them on Mainnet or any other live network WILL BE LOST."
