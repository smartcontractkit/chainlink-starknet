#!/usr/bin/env bash
set -euxo pipefail

cd contracts
yarn install
# Remove once https://github.com/Shard-Labs/starknet-hardhat-plugin/pull/106 is merged
npx hardhat starknet-compile
yarn test
# Example tests
cd examples/ocr2
yarn compile && yarn test

# Validator tests
cd ../../../emergency-protocol
yarn compile && yarn compile:l1 && yarn test
