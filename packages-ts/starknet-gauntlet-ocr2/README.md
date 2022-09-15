# Starknet Gauntlet Commands for Chainlink OCR2 Protocol

## Set up

Make sure you have your account details set up in your `.env` file

```bash
# .env
PRIVATE_KEY=0x...
ACCOUNT=0x...
LINK=0x...
```

Note: The [token contract](https://github.com/smartcontractkit/chainlink-starknet/tree/develop/packages-ts/starknet-gauntlet-starkgate) should only be deployed once and the same contract should be used for very aggregator

## Deploy an Access Controller Contract

Run the following command:

```bash
yarn gauntlet access_controller:deploy --network=<NETWORK>
```

This command will generate a new Access Controller address and will give the details during the deployment. You can then add this contract address in your `.env` files as `BILLING_ACCESS_CONTROLLER`, or pass it into the command directly (as shown below)

## Deploy an OCR2 Contract

Run the following command substituting the following attributes:

1. <TOKEN_CONTRACT> with the [deployed token contract](https://github.com/smartcontractkit/chainlink-starknet/tree/develop/packages-ts/starknet-gauntlet-starkgate)
2. <ACCESS_CONTROLLER_CONTRACT> with the access controller contract you deployed in the previous section

```bash
yarn gauntlet ocr2:deploy --network=<NETWORK> --billingAccessController=<ACCESS_CONTROLLER_CONTRACT> --minSubmissionValue=<MIN_VALUE> --maxSubmissionValue=<MAX_VALUE> --decimals=<DECIMALS> --name=<FEED_NAME> --link=<TOKEN_CONTRACT>
```

This command will generate a new OCR2 address and will give the details during the deployment

## Set the Billing Details on OCR2 Contract

Run the following command substituting <OCR_CONTRACT_ADDRESS> with the OCR2 contract address you received in the deploy step:

```
yarn gauntlet ocr2:set_billing --observationPaymentGjuels=<AMOUNT> --transmissionPaymentGjuels=<AMOUNT> <CONTRACT_ADDRESS>
```

This Should set the billing details for this feed on contract address

## Set the Config Details on OCR2 Contract

Run the following command substituting <OCR_CONTRACT_ADDRESS> with the OCR2 contract address you received in the deploy step:

```
yarn gauntlet ocr2:set_config --network=<NETWORK> --address=<ADDRESS> --f=<NUMBER> --signers=[<ACCOUNTS>] --transmitters=[<ACCOUNTS>] --onchainConfig=<CONFIG> --offchainConfig=<CONFIG> --offchainConfigVersion=<NUMBER> <OCR_CONTRACT_ADDRESS>
```

This Should set the config for this feed on contract address.
