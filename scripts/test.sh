#!/usr/bin/env bash
set -euxo pipefail

sh ./integration-tests/scripts/devnet-hardhat.sh
cd contracts
yarn install
yarn compile:l1 && yarn compile
yarn test
# Example tests
cd ../examples/contracts/aggregator-consumer
yarn install
yarn compile && yarn test
