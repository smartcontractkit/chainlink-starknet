This example demonstrates how consumer contracts can read from the aggregator, with and without consideration of the uptime feed address. Recall that the purpose of the uptime feed is to let consumer contracts know if the L2 layer is up and healthy.

We demonstrate these examples by mocking an aggregator to return stubbed values.

You can run this example on your local devnet or on goerli (up to you). NOTE: at the time of writing, the examples only work against devnet due to Starknet version incompatability.

Note: `.env` will contain account details, such as DEPLOYER_ACCOUNT_ADDRESS and DEPLOYER_PRIVATE_KEY

## Deploy on Devnet

### 1. Install Local Dev Environment

Follow [steps 1 and 2 here](../proxy_consumer/README.md)

At this point, you should have starknet-devnet, cairo, and scarb installed. Your virtualenv will also have the starknet cli tool but we won't be using that to test against devnet

All instructions begin at the root of this folder
### 2. Compile cairo 1 contracts

```
    yarn compile:cairo
```


### 3. Setup Starknet Devnet

In a seperate terminal, run:
```
    starknet-devnet --cairo-compiler-manifest ../../vendor/cairo/Cargo.toml --seed 0 --lite-mode ../../../vendor/cairo/Cargo.toml 
```
This will start up the devnet and enable you to deploy and run cairo 1 contracts against it

### 4. Deploy Devnet Account

This command will deploy a devnet account onto devnet and write the DEPLOYER_ACCOUNT_ADDRESS and DEPLOYER_PRIVATE_KEY into a .env file in the current directory (it will create .env if it doesn't exist). This account will be utilized for deploying and interacting with the rest of the contracts in this section.

```
    yarn deployAccount
```

## 5. Deploy Contracts

This will deploy the following contracts:
* SequencerUptimeFeed: Displays the status of the L2 layer. If it is up, then it is safe to read from the aggregator contract. For more information please see this [document](../../../docs/emergency-protocol/README.md). 
* MockAggregator: A mocked version of the aggregator with limited functionality. It gives the reader the ability to set and view the latest round data in order so the reader can familiarize themselves for testing purposes.
* AggregatorConsumer: Simply reads the mocked aggregator's latest values
* AggregatorPriceConsumerWithSequencer: Reads the mocked aggregator's values but also queries the SequencerUptimeFeed to determine if the value should be used or not. If the SequencerUptimeFeed is too stale, then that means the L2 layer is down. We've arbitrarily chosen the threshold of 60 seconds.


```
    yarn deployContracts
```

## 6. Interact with Contracts

```
# deployer calls AggregatorConsumer to read the decimals method of the MockAggregator
yarn readDecimals

# deployer calls AggregatorConsumer to read the latest round of the MockAggregator
yarn readLatestRound

# deployer calls Aggregator Consumer to poll decimals AND latest round of MockAggregator 
yarn readContinuously

# deployer calls MockAggregator to manually set the new round's data
yarn updateLatestRound

# deployer calls AggregatorPriceConsumerWithSequencer to read latest round or revert if uptime feed is stale
yarn getLatestPriceSeqCheck
