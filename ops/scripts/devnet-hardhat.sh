#!/usr/bin/env bash
# TODO: this script needs to be replaced with a predefined K8s enviroment

set -euo pipefail

cpu_struct=`arch`;
echo $cpu_struct;

git_root="$(git rev-parse --show-toplevel)"
cairo_path="${git_root}/vendor/cairo"
cairo_compiler_manifest="${cairo_path}/Cargo.toml"

if [ ! -f "${cairo_compiler_manifest}" ]; then
  echo "No Cargo.toml; did you checkout the cairo git submodule?"
  exit 1
fi

node --version;

# Clean up first
bash "$(dirname -- "$0";)/devnet-hardhat-down.sh"

echo "Checking CPU structure..."
if [[ $cpu_struct == *"arm"* ]]
then
    echo "Starting arm devnet container..."
    container_version="0.5.1-arm"
else
    echo "Starting i386 devnet container..."
    container_version="0.5.1"
fi

# we need to replace the entrypoint because starknet-devnet's docker builds at 0.5.1 don't include cargo or gcc.
docker run \
  -p 127.0.0.1:5050:5050 \
  -p 127.0.0.1:8545:8545 \
  -d \
  --name chainlink-starknet.starknet-devnet \
  --volume "${cairo_path}:/cairo" \
  --entrypoint sh \
  "shardlabs/starknet-devnet:${container_version}" \
  -c '(wget https://sh.rustup.rs -O - | sh -s -- -y) && apk add gmp-dev g++ gcc libffi-dev && PATH="/root/.cargo/bin:${PATH}" starknet-devnet --lite-mode --host 0.0.0.0 --cairo-compiler-manifest /cairo/Cargo.toml'

echo "Starting hardhat..."
docker run --net container:chainlink-starknet.starknet-devnet -d --name chainlink-starknet.hardhat ethereumoptimism/hardhat-node:nightly

# starknet-devnet startup is slow and requires compiling cairo.
echo "Waiting for starknet-devnet to become ready.."
start_time=$(date +%s)
while true
do
  output=$(docker logs chainlink-starknet.starknet-devnet 2>&1)

  if [[ $output == *"Listening"* ]]; then
    echo "${output}"
    echo ""
    echo "starknet-devnet is ready."
    exit 0
  fi

  current_time=$(date +%s)
  elapsed_time=$((current_time - start_time))

  if (( elapsed_time > 600 )); then
    echo "Error: Command did not become ready within 600 seconds"
    exit 1
  fi

  sleep 3
done
