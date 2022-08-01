#!/usr/bin/env bash
set -euxo pipefail

cd contracts
npm install
# Remove once https://github.com/Shard-Labs/starknet-hardhat-plugin/pull/106 is merged
npx hardhat starknet-compile
yarn test
# Example tests
cd ./examples/contracts/aggregator-consumer
yarn install
yarn compile && yarn test

# Validator tests
cd ../../..
sh ./integration-tests/scripts/devnet-hardhat.sh
