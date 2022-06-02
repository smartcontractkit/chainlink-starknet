# Gauntlet Starknet CLI

This packages expose the commands to be used as a CLI of the following packages:

- @chainlink/gauntlet-starknet-example
- @chainlink/gauntlet-starknet-account
- @chainlink/gauntlet-starknet-ocr2

## Setup

Every command accepts the `--network=<NETWORK>` flag. The value will load the static environment variables under `./networks/.env.<NETWORK>`. Currently 2 network configurations are available:

1. Local

```bash
NODE_URL=http://127.0.0.1:5000
```

2. Testnet

```bash
NODE_URL=https://alpha4.starknet.io
```

## Miscellaneous

- To show available methods for all contracts:

```
yarn gauntlet -h
```

-To show available methods for a single contract:

```
yarn gauntlet ocr2:deploy -h
```

It will show details for the specificed function with their needed parameters and their types, if any
