This is a simple example for how to read Chainlink data feeds on Starknet.

### Requirements

Set up your environment to run the examples. Make sure to clone this repo before you start these instructions

1. Clone the [smartcontractkit/chainlink-starknet](https://github.com/smartcontractkit/chainlink-starknet) repository, which includes the example contracts for this guide: 
    ```
    git clone https://github.com/smartcontractkit chainlink-starknet.git
    git submodule update --init --recursive
    ```
    We use git submodules to pin specific versions of cairo and scarb (you'll see this come into play later).

1. Setup your local Starknet environment. We will install starknet cli, the rust cairo compiler, and scarb which is a framework and dependency manager for cairo. If you already have them installed, feel free to skip this step (if you later find that your versions are not working, follow the steps below because they are pinned to specific versions).
    ```
    # Part 1: Install starknet cli via virtualenv
    
    cd chainlink-starknet
    # tested on python 3.9 and onwards
    python -m venv venv 
    source ./venv/bin/activate
    pip install -r contracts/requirements.txt
    ```
    Next we'll install cairo. If you've already installed cairo, make sure to disable that path first.
    ```
        # Part 2: Install cairo 
        cd vendor/cairo && cargo build --all --release
        # Add cairo executable to your path
        export PATH="$HOME/path/to/chainlink-starknet/vendor/cairo/target/release:$PATH"  
    ``` 
    Lastly, we'll install scarb. You should be able to install the scarb 0.2.0-alpha.2 binary from [here](https://github.com/software-mansion/scarb/releases/tag/v0.2.0-alpha.2) for your operating system. Install it in your `$HOME` directory and add it to your path.
    ```
        # assuming you've downloaded the scarb artifact already to your $HOME directory
        export PATH="$HOME/scarb/bin:$PATH"
    ```
    Awesome, that was a lot of work, but now we're ready to start!

1. Set up a Starknet account. Follow instructions to [set up environment variables](https://docs.starknet.io/documentation/getting_started/deploying_contracts/#setting_up_environment_variables) and [deploy an account](https://docs.starknet.io/documentation/getting_started/deploying_contracts/#setting_up_an_account). This deploys an account on Starknet's `alpha-goerli` network and funds it with [testnet ETH](https://faucet.goerli.starknet.io/). These examples expect the OpenZeppelin wallet, which stores your addresses and private keys at `~/.starknet_accounts/` by default.

1. [Install NodeJS](https://nodejs.org/en/download/) in the version in the `>=14 <=18` version range.
1. [Install Yarn](https://classic.yarnpkg.com/lang/en/docs/install/).
1. Change directories to the proxy consumer example: `cd ./chainlink-starknet/examples/new_contracts/proxy_consumer/`
1. Run `yarn install` to install the required packages including [Starknet.js](https://www.starknetjs.com/)

### Running the on-chain example

1. Find the your account address and private key for your funded Starknet testnet account. By default, the OpenZeppelin wallet contains these values at `~/.starknet_accounts/starknet_open_zeppelin_accounts.json`.
1. Export your address to the `DEPLOYER_ACCOUNT_ADDRESS` environment variable and your private key to the `DEPLOYER_PRIVATE_KEY` environment variable.

   ```shell
   export DEPLOYER_ACCOUNT_ADDRESS=<YOUR_WALLET_ADDRESS>
   ```

   ```shell
   export DEPLOYER_PRIVATE_KEY=<YOUR_KEY>
   ```
1. Run `yarn build` to build the cairo artifacts via scarb. These will be put in the target/ directory
1. Run `yarn deploy` to deploy the example consumer contract to the Starknet Goerli testnet. The console prints the contract address and transaction hash.
1. Run `yarn readLatestRound <CONTRACT_ADDRESS>` to send an invoke transaction to the deployed contract. Specify the contract address printed by the deploy step. The deployed contract reads the latest round data from the proxy, stores the values, and prints the resulting values.

### Running the off-chain example

This example simply reads the proxy contract to get the latest values with no account or contract compiling steps required.

1. Run `yarn readLatestRoundOffChain`.
