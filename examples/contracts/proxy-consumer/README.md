This is a simple example for how to read Chainlink data feeds on Starknet.

### Requirements

- Install [Python 3.9](https://www.python.org/downloads/).
- Install [Yarn](https://classic.yarnpkg.com/lang/en/docs/install/).
- [Setup your local StarkNet environment](https://starknet.io/docs/quickstart.html).
- [Set up a StarkNet account](https://starknet.io/docs/hello_starknet/account_setup.html) on Starknet's `alpha-goerli` network and fund it with [testnet ETH](https://faucet.goerli.starknet.io/).

### Running the on-chain example

1. Run `yarn compile` to create the `./starknet-artifacts/`.
1. Run `yarn deploy` to deploy the example consumer contract to the StarkNet Goerli testnet. The console prints the contract address and transaction hash.
1. Run `yarn readLatestRound <CONTRACT_ADDRESS>` to send an invoke transaction to the deployed contract. Specify the contract address printed by the deploy step. The deployed contract reads the latest round data from the proxy, stores the values, and prints the resulting values.

### Running the off-chain example

This example simply reads the proxy contract to get the latest values with no account or contract compiling steps required.

1. Run `yarn readLatestRoundOffChain`.
