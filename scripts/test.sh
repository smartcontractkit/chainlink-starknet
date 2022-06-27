#!/usr/bin/env bash
set -euxo pipefail

cd contracts
yarn install
# Remove once https://github.com/Shard-Labs/starknet-hardhat-plugin/pull/106 is merged
npx hardhat starknet-compile
yarn test
# Example tests
cd contracts/examples/ocr2
yarn test
