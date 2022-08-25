#!/usr/bin/env bash
set -euxo pipefail

# TODO: this script needs to be replaced with a predefined K8s enviroment
sh ./ops/scripts/devnet-hardhat.sh

cd contracts
yarn install
yarn compile
yarn test

# Example tests
cd ../examples/contracts/aggregator-consumer
yarn install
yarn compile
yarn test
