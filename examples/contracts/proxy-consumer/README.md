This is a simple example for how to read Chainlink data feeds on Starknet.

### Requirements

Set up your environment to run the examples.

1. [Setup your local StarkNet environment](https://starknet.io/docs/quickstart.html). Note that a Python version in the `>=3.6 <=3.9` range is required for compiling and deploying contracts on-chain. The [`cairo-lang` Python package](https://pypi.org/project/cairo-lang/) is not compatible with newer versions of Python as of the [`cairo-lang` 0.10.3](https://pypi.org/project/cairo-lang/0.10.3/) package. Check [starknet.io](https://starknet.io/docs/quickstart.html) for the latest requirements.
1. [Set up a StarkNet account](https://starknet.io/docs/hello_starknet/account_setup.html) on StarkNet's `alpha-goerli` network and fund it with [testnet ETH](https://faucet.goerli.starknet.io/). These examples expect the OpenZeppelin wallet, which stores your addresses and private keys at `~/.starknet_accounts/starknet_open_zeppelin_accounts.json` by default.
1. [Install NodeJS](https://nodejs.org/en/download/) in the version in the `>=14 <=18` version range.
1. [Install Yarn](https://classic.yarnpkg.com/lang/en/docs/install/).
1. Clone the [smartcontractkit/chainlink-starknet](https://github.com/smartcontractkit/chainlink-starknet) repository, which includes the example contracts for this guide: `git clone https://github.com/smartcontractkit/chainlink-starknet.git`
1. In your clone of the [chainlink-starknet](https://github.com/smartcontractkit/chainlink-starknet) repository, change directories to the proxy consumer example: `cd ./chainlink-starknet/examples/contracts/proxy-consumer/`
1. Run `yarn install` to install the required packages including [StarkNet.js](https://www.starknetjs.com/), [HardHat](https://hardhat.org/), and the [StarkNet Hardhat Plugin](https://shard-labs.github.io/starknet-hardhat-plugin/).

### Running the on-chain example

1. Find the your account address and private key for your funded StarkNet testnet account. By default, the OpenZeppelin wallet contains these values at `~/.starknet_accounts/starknet_open_zeppelin_accounts.json`.
1. Export your address to the `DEPLOYER_ACCOUNT_ADDRESS` environment variable and your private key to the `DEPLOYER_PRIVATE_KEY` environment variable.

    ```shell
    export DEPLOYER_ACCOUNT_ADDRESS=<YOUR_WALLET_ADDRESS>
    ```

    ```shell
    export DEPLOYER_PRIVATE_KEY=<YOUR_KEY>
    ```

1. Run `yarn build` to run Hardhat and create `./starknet-artifacts/` with the compiled contracts. Hardhat uses the [`@shardlabs/starknet-hardhat-plugin` package](https://www.npmjs.com/package/@shardlabs/starknet-hardhat-plugin) for this step.
1. Run `yarn deploy` to deploy the example consumer contract to the StarkNet Goerli testnet. The console prints the contract address and transaction hash.
1. Run `yarn readLatestRound <CONTRACT_ADDRESS>` to send an invoke transaction to the deployed contract. Specify the contract address printed by the deploy step. The deployed contract reads the latest round data from the proxy, stores the values, and prints the resulting values.

### Running the off-chain example

This example simply reads the proxy contract to get the latest values with no account or contract compiling steps required.

1. Run `yarn readLatestRoundOffChain`.
