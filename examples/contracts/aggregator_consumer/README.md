# Examples

## Overview

In this directory you'll find three top-level folders:

- `src/`: contains sample cairo contracts that demonstrate how one can integrate with Chainlink's core starknet contracts.
- `tests/`: contains cairo tests for the example contracts in `src/`. They showcase some simple usage patterns for the contracts.
- `scripts/`: contains cairo scripts that allow you to interact with the example contracts over testnet or a local starknet devnet container. 

## Prerequisites

To get started, ensure that you have the following tools installed on your machine:

- [starknet-foundry (v0.20.1)](https://github.com/foundry-rs/starknet-foundry/releases/tag/v0.18.0)
- [scarb (v2.5.4)](https://github.com/software-mansion/scarb/releases/tag/v2.5.4)

## Tests

To run all test cases in the `tests/` directory, you can use the following command:

```sh
snforge test
```

## Scripts

### Setup

#### Using a Local Starknet Docker Devnet

If you would like to run the scripts against a local starknet devnet container:

- First, execute the following command to run a [starknet-devnet-rs](https://github.com/0xSpaceShard/starknet-devnet-rs) container:

  ```sh
  make devnet
  ```

  If this command is re-run, it will fully stop the container and recreate it. This can be useful in case you'd like to restart from a completely clean state and re-run the deploy scripts.

- Next, make sure that the `url` property in the `snfoundry.toml` file is configured such that it points to the docker container:

  ```toml
  [sncast.default]
  url = "http://127.0.0.1:5050/rpc"
  ```

- The starknet devnet container comes with a set of prefunded accounts. In order to run the scripts, we'll need to add one of these accounts to our local accounts file (usually located at `~/.starknet_accounts/starknet_open_zeppelin_accounts.json`). This can be done using the following command:

  ```sh
  make add-account
  ```

- If you would like to view the account info later in the future, you can run:

  ```sh
  make view-accounts
  ```

At this point you should be ready to start executing scripts! Feel free to move onto the next section.

#### Using Testnet

If you would like to run the scripts against testnet:

- First, make sure that the `url` property in the `snfoundry.toml` file is configured such that it points to Starknet testnet:

  ```toml
  [sncast.default]
  url = "https://starknet-sepolia.public.blastapi.io/rpc/v0_7"
  ```

- Next, we'll need to create an account. This account will be used to pay for testnet transaction fees for some of the scripts. To generate an account (but not necessarily deploy it to the network yet), you can run the following command:

  ```sh
  make create-account
  ```
  
- Once the account has been created, its info should be stored in an accounts file on your local machine (usually at `~/.starknet_accounts/starknet_open_zeppelin_accounts.json`). If you'd like to view the account information later on, you can use the following command:

  ```sh
  make view-accounts
  ```

- Next, you'll need to fund the account with some tokens. This can be achieved by sending tokens from another starknet account or by bridging them with [StarkGate](https://sepolia.starkgate.starknet.io).

- After you fund the account, you can deploy it to testnet using the following command:

  ```sh
  make deploy-account
  ```

At this point you should be ready to start executing scripts! Feel free to move onto the next section.

### Running Scripts

#### Test Consumer

In this section, we'll deploy a `MockAggregator` contract and a `AggregatorConsumer` contract to the network.

- To deploy the contracts to the network, you can use the following command:

  ```sh
  make tc-deploy
  ```

  Under the hood, this command runs a declare transaction followed by a deploy transaction for each contract. It outputs the transaction hashes as well as the contract addresses (caution - the addresses are not hex encoded). If you are using testnet, be sure to note the contract addresses and convert them to hex as they will be needed in later steps. If you are using a local devnet container, this command can only be run once - re-running it will lead to a transaciton error. In order to re-run the deploy script you'll need to re-run the `make devnet` command as mentioned above. 

- Once the contracts are deployed, you can start calling methods on them! If you're using a local devnet, the contract addresses in the scripts should already be correct, and you can move onto the following steps. However, if you're using testnet, you'll need to make sure that the contract addresses in each of the following scripts are modified accordingly. 

- We'll start by calling the `read_decimals` function on the `AggregatorConsumer`:
  
  ```sh
  make tc-read-decimals
  ```

  This should return an output similar to:

  ```text
  Result::Ok(CallResult { data: [16] })
  command: script run
  status: success
  ```

- To read the latest round data from the `AggregatorConsumer`, you can use the following command:

  ```sh
  make tc-read-latest-round
  ```

  This should return an output like:

  ```text
  Result::Ok(CallResult { data: [0, 0, 0, 0, 0] })
  command: script run
  status: success
  ```

  In this case, there isn't any round data yet since we haven't added any mock data. Let's do that right now!

- To set the latest round data, you can run the following command:
  
  ```sh
  make tc-set-latest-round 
  ```

  This should result in an output like:

  ```text
  Transaction hash = 0x5b57df1db0898caefd01f7d7ff9a300814ef8869a6c475f135d8b5d56e0e3a8
  Result::Ok(InvokeResult { transaction_hash: 2582232800348643522264958893576302212545891688073192089151947336582678242216 })
  command: script run
  status: success
  ```

  Under the hood, this script sends a transaction to the network which calls `set_latest_round_data` on the `MockAggregator`. The data sent to the function is hardcoded in the script and can be modified in any way you like.

- Now when we read the latest round data:

  ```sh
  make tc-latest-round
  ```

  We should see something like:

  ```text
  Result::Ok(CallResult { data: [1, 1, 12345, 100000, 200000] })
  command: script run
  status: success
  ```

  This array of values represents the following:

  ```text
  [   1,      1,      12345,           100000,                200000        ]
  [roundId, answer, block_num, observation_timestamp, transmittion_timestamp]
  ```

At this point you should be familiar with running the example scripts and transactions!
