#!/usr/bin/env bash

set -euo pipefail

bash "$(dirname -- "$0")/core.down.sh"

container_name_prefix="chainlink.core"
container_version="starknet"

# https://github.com/smartcontractkit/chainlink/blob/600365a7a27508d699dbd4b994fafba7dc288659/integration-tests/client/chainlink_k8s.go#L82-L83
api_email="notreal@fakeemail.ch"
api_password="fj293fbBnlQ!f9vNs"

if [[ -z "${CL_CONFIG:-}" ]]; then
	echo "No CL_CONFIG env var provided." >&2
	exit 1
fi

platform_arg=""
if [ -n "${CORE_IMAGE:-}" ]; then
	image_name="${CORE_IMAGE}"
else
	image_name="smartcontract/chainlink:${container_version}"
	platform_arg="--platform linux/amd64"
fi
echo "Using core image: ${image_name}"

listen_ips="0.0.0.0"

NODE_COUNT="${NODE_COUNT:-5}"

declare -i core_base_port=50100
declare -i core_p2p_base_port=50200

for ((i = 1; i <= NODE_COUNT; i++)); do
	database_name="core_test_${i}"
	echo "Creating database: ${database_name}"
	# postgres_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' chainlink.postgres)
	# TODO: use postgres db ip
	host_postgres_url="postgresql://postgres:postgres@127.0.0.1:5432/postgres"
	psql "${host_postgres_url}" -c "DROP DATABASE ${database_name};" &>/dev/null || true
	psql "${host_postgres_url}" -c "CREATE DATABASE ${database_name};" &>/dev/null || true

	# TODO: remove this and use node ids
	listen_args=()
	for ip in $listen_ips; do
		listen_args+=("-p" "${ip}:$((core_base_port + i - 1)):6688")
		# listen_args+=("-p" "${ip}:$((core_p2p_base_port + i - 1)):6690")
	done

	echo "Starting core container $i"
	container_name="${container_name_prefix}.$i"

		# --rm \
	docker run \
		-d \
		$platform_arg \
		"${listen_args[@]}" \
		--add-host=host.containers.internal:host-gateway \
		--name "${container_name}" \
		--network-alias "${container_name}" \
		--network chainlink \
		-e "CL_CONFIG=${CL_CONFIG}" \
		-e "CL_DATABASE_URL=postgresql://postgres:postgres@chainlink.postgres:5432/${database_name}?sslmode=disable" \
		-e 'CL_PASSWORD_KEYSTORE=asdfasdfasdfasdf' \
		--entrypoint bash \
		"${image_name}" \
		-c \
		"echo -e \"${api_email}\\n${api_password}\" > /tmp/api_credentials && chainlink node start --api /tmp/api_credentials"

	echo "Waiting for core container to become ready.."
	start_time=$(date +%s)
	prev_output=""
	while true; do
		output=$(docker logs "${container_name}" 2>&1)
		if [[ "${output}" != "${prev_output}" ]]; then
			echo -n "${output#$prev_output}"
			prev_output="${output}"
		fi

		if [[ $output == *"Listening and serving HTTP"* ]]; then
			echo ""
			echo "node is ready."
			break
		fi

		current_time=$(date +%s)
		elapsed_time=$((current_time - start_time))

		if ((elapsed_time > 600)); then
			echo "Error: Command did not become ready within 600 seconds"
			exit 1
		fi

		sleep 3
	done
done
