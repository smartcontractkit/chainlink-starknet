This example demonstrates how consumer contracts can read from the aggregator, with and without consideration of the uptime feed address. Recall that the purpose of the uptime feed is to let consumer contracts know if the L2 layer is up and healthy.

We demonstrate these examples by mocking an aggregator to return stubbed values.

You can run this example on your local devnet or on goerli (up to you).

Note: `.env` will contain account details, such as DEPLOYER_ACCOUNT_ADDRESS and DEPLOYER_PRIVATE_KEY

## Deploy on Devnet

### 1. Install and run starknet-devnet

```
pip install starknet-devnet
starknet-devnet
```

At this point, 

### 2. 



## Create Starknet Account

1. Set up the network `export STARKNET_NETWORK=alpha-goerli`

2. Choose a wallet provider `export STARKNET_WALLET=starkware.starknet.wallets.open_zeppelin.OpenZeppelinAccount`

3. If you don't have an account create it with the command `starknet new_account`. If you do not have the cli installed, see https://docs.starknet.io/documentation/getting_started/environment_setup/ on how to set up the starknet cli

4. Open that link https://faucet.goerli.starknet.io/ and past the address to get some faucet. If you can not get faucet you can transfer some eth from one of you L1 accounts to starknet https://goerli.starkgate.starknet.io/.

5. When you have been able to get some faucet you can deploy your account `starknet deploy_account` (for more information take a look at this documentation https://starknet.io/docs/hello_starknet/account_setup.html).

## Set environment variables

6. Create a `.env` in `examples/contracts/aggregator-consumer`and copy paste your account address under `DEPLOYER_ACCOUNT_ADDRESS` and you private key under `DEPLOYER_PRIVATE_KEY`. You can find the private key in your home directory under `.starknet_accounts` folder in `starknet_open_zeppelin_accounts.json`.

## Deploy Contracts

7. Deploy the contracts by running `npx ts-node ./scripts/deploy_contracts.ts`. Set environment variable NETWORK=GOERLI to run this against starknet goerli.

## Interact with Contracts

8. You can now read decimals and latest round data by running `yarn readDecimals` and `yarn readLatestRound`.

9. You can update value and see continuously the data by openning another terminal and run `npx ts-node ./scripts/updateLatestRound.ts` in one terminal and `npx ts-node ./scripts/readContinuously` into another one.
