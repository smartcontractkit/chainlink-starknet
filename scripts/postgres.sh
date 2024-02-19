#!/usr/bin/env bash

container_name="chainlink.postgres"
container_version="15.2-alpine"

set -euo pipefail

bash "$(dirname -- "$0")/postgres.down.sh"

listen_ips="0.0.0.0"

echo "Starting postgres container"

listen_args=()
for ip in $listen_ips; do
	listen_args+=("-p" "${ip}:5432:5432")
done

docker run \
	"${listen_args[@]}" \
	-d \
	--name "${container_name}" \
	--network-alias "${container_name}" \
	--network chainlink \
	-e POSTGRES_USER=postgres \
	-e POSTGRES_PASSWORD=postgres \
	-e POSTGRES_DB=chainlink_test \
	"postgres:${container_version}" \
	-c 'listen_addresses=*'

echo "Waiting for postgres container to become ready.."
start_time=$(date +%s)
prev_output=""
while true; do
	output=$(docker logs "${container_name}" 2>&1)
	if [[ "${output}" != "${prev_output}" ]]; then
		echo -n "${output#$prev_output}"
		prev_output="${output}"
	fi

	if [[ $output == *"listening on IPv4 address"* ]]; then
		echo ""
		echo "postgres is ready."
		exit 0
	fi

	current_time=$(date +%s)
	elapsed_time=$((current_time - start_time))

	if ((elapsed_time > 600)); then
		echo "Error: Command did not become ready within 600 seconds"
		exit 1
	fi

	sleep 3
done
