# Examples

## Overview

In this directory you'll find three top-level folders:

- `src/`: contains sample cairo contracts that demonstrate how one can integrate with Chainlink's core starknet contracts.
- `tests/`: contains cairo tests for the example contracts in `src/`. They showcase some simple usage patterns for the contracts.
- `scripts/`: contains cairo scripts that allow you to interact with the example contracts over testnet or a local starknet devnet container. 

## Prerequisites

To get started, ensure that you have the following tools installed on your machine:

- [starknet-foundry (v0.20.1)](https://github.com/foundry-rs/starknet-foundry/releases/tag/v0.20.1)
- [scarb (v2.5.4)](https://github.com/software-mansion/scarb/releases/tag/v2.5.4)

## Tests

To run all test cases in the `tests/` directory, you can use the following command:

```sh
make test
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

- The starknet devnet container comes with a set of prefunded accounts. In order to run the scripts, we'll need to add one of these accounts to a local `accounts.json` file. This can be done using the following command:

  ```sh
  make add-account
  ```

At this point you should be ready to start executing scripts! Feel free to move onto the next section.

#### Using Testnet

If you would like to run the scripts against testnet:

- First, let's generate our account details. We can do this by running the following command:

  ```sh
  make create-account
  ```
  
  Once the account has been created, its info should be stored in an accounts file on your local machine (usually at `~/.starknet_accounts/starknet_open_zeppelin_accounts.json`). 

- Next, you'll need to fund the account with some tokens. This can be achieved by sending tokens from another starknet account or by bridging them with [StarkGate](https://sepolia.starkgate.starknet.io).

- After you fund the account, you can deploy it to testnet using the following command:

  ```sh
  make deploy-account
  ```

At this point you should be ready to start executing scripts! Feel free to move onto the next section.

### Running Scripts

There are several different ways to use the scripts in this repo. We'll cover a few different options below.

#### Reading Data from an Aggregator

##### Devnet

First, let's deploy a mock aggregator contract to our container:

```sh
make ma-deploy NETWORK=devnet
```

Under the hood this command will run a declare transaction followed by a deploy transaction for the MockAggregator contract. This command should output something similar to:

```text
Declaring and deploying MockAggregator
Declaring contract...
Transaction hash = 0x568d29d07128cba750845b57a4bb77a31f628b6f4288861d8b31d12e71e4c3b
Class hash = 301563338814178704943249302673347019225052832575378055777678731916437560881
Deploying contract...
Transaction hash = 0xfbc49eb82894a704ce536ab904cdee0fd021b0fba335900f8b9b12cfcd005f
MockAggregator deployed at address: 1566652744716179301065270359129119857774335542042051464747302084192731701184

command: script run
status: success
```

Once the MockAggregator is deployed, you can read the latest round data using the following command:

```sh
make agg-read-latest-round NETWORK=devnet
```

This should return an output like:

```text
Result::Ok(CallResult { data: [0, 0, 0, 0, 0] })
command: script run
status: success
```

In this case, there isn't any round data yet since we haven't added any mock data. To set the latest round data, you can run the following command:

```sh
make ma-set-latest-round NETWORK=devnet
```

This should result in an output like:

```text
Transaction hash = 0x5b57df1db0898caefd01f7d7ff9a300814ef8869a6c475f135d8b5d56e0e3a8
Result::Ok(InvokeResult { transaction_hash: 2582232800348643522264958893576302212545891688073192089151947336582678242216 })
command: script run
status: success
```

Under the hood, this script sends a transaction to the network which calls `set_latest_round_data` on the `MockAggregator`. The data sent to the function is hardcoded in the script and can be modified in any way you like. Now when we read the latest round data:

```sh
make agg-read-latest-round NETWORK=devnet
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

##### Testnet

The steps and commands used for devnet can also be applied to testnet! However, there are a few noticeable differences:

- You do not need to deploy a mock aggregator. If you already have the address of a pre-deployed aggregator, you can use it in the `read_latest_round.cairo` script!

- For all the commands, make sure you use `NETWORK=testnet`.

#### Deploying an Aggregator Consumer Contract

##### Devnet

First, let's restart our devnet container to ensure we're starting from a clean slate:

```sh
make devnet
```

Once the container is restarted, let's deploy the MockAggregator contract and the AggregatorConsumer contract to it:

```sh
make devnet-deploy
```

The AggregatorConsumer takes the address of an aggregator contract as input (in this case it is the MockAggregator). It comes with two methods:

- `set_answer`: this function reads the latest round data from the aggregator and stores the round answer in storage.
- `read_answer`: this function reads the answer from storage. The answer is initially set to 0 on deployment.

At this point, the latest round data has not been set, so calling `set_answer` on the AggregatorConsumer contract will trivially set the answer to 0. You can run the following commands to verify this:

- Let's check that the answer is initially set to 0:

  Command:

  ```sh
  make ac-read-answer NETWORK=devnet
  ```

  Output:


  ```text
  Result::Ok(CallResult { data: [0] })
  command: script run
  status: success
  ```

- Let's set the answer:

  Command:

  ```sh
  make ac-set-answer NETWORK=devnet 
  ```

  Output:

  ```text
  Transaction hash = 0x7897b80451a4e4d6df1dc575fffe9b6ebc774bd675eb24c8cf83e0a3818071
  Result::Ok(InvokeResult { transaction_hash: 213068772556793858646692905972104002796353690311811440814152767066255491185 })
  command: script run
  status: success
  ```

- The answer should still be 0:

  Command:

  ```sh
  make ac-read-answer NETWORK=devnet
  ```

  Output:


  ```text
  Result::Ok(CallResult { data: [0] })
  command: script run
  status: success
  ```


To change this, let's set the latest round data to some dummy values:

```sh
make ma-set-latest-round NETWORK=devnet 
```

Then let's refresh the AggregatorConsumer's answer:

```sh
make ac-set-answer NETWORK=devnet 
```

Finally, let's read the AggregatorConsumer's answer:

```sh
make ac-read-answer NETWORK=devnet
```

This should result in the following:

```text
Result::Ok(CallResult { data: [1] })
command: script run
status: success
```

##### Testnet

The steps and commands used for devnet can also be applied to testnet! However, there are a few noticeable differences:

- You do not need to use the `make devnet` command or the `make devnet-deploy` command. For contract deployment, you'll most likely want to use the address of a pre-deployed aggregator instead of the MockAggregator. If this is the case, you won't have control over the latest round data, so you can ignore the commands that interact with the MockAggregator (i.e. `make ma-set-latest-round`). For deployment, you can perform the following:
  1. note the address of the Aggregator contract you'd like to use
  1. input the address in the `deploy_aggregator_consumer.cairo` script
  1. run `make ac-deploy NETWORK=testnet`

- For all the commands, make sure you use `NETWORK=testnet`.

- The AggregatorConsumer scripts (e.g. `read_answer.cairo` and `set_answer.cairo`) contain the hardcoded address of the AggregatorConsumer for devnet. If you'd like to use these scripts on testnet, the hardcoded AggregatorConsumer address in these scripts will need to be swapped with the address that you receive from `make ac-deploy`. Keep in mind that the address that is printed from `make ac-deploy` may not be hex encoded - if this is the case you'll need to convert it to hex before adding it to the script. 

