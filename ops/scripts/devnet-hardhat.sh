#!/usr/bin/env bash
# TODO: this script needs to be replaced with a predefined K8s enviroment

cpu_struct=`arch`;
echo $cpu_struct;

node --version;

# Clean up first
bash "$(dirname -- "$0";)/devnet-hardhat-down.sh"

echo "Checking CPU structure..."
if [[ $cpu_struct == *"arm"* ]]
then
    echo "Starting arm devnet container..."
    docker run -p 5050:5050 -p 8545:8545 -d --name chainlink-starknet.starknet-devnet shardlabs/starknet-devnet:0.5.0;
else
    echo "Starting i386 devnet container..."
    docker run -p 5050:5050 -p 8545:8545 -d --name chainlink-starknet.starknet-devnet shardlabs/starknet-devnet:0.5.0;
fi

echo "Starting hardhat..."
docker run --net container:chainlink-starknet.starknet-devnet -d --name chainlink-starknet.hardhat ethereumoptimism/hardhat-node:nightly
