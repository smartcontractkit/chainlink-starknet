# Getting started

## Setup

Make sure you have Node and Yarn installed

Node: https://nodejs.org/es/download/

Yarn:

```
npm install --global yarn
```

### Install

```
yarn
```

### Run

To see the available commands, run:

```
yarn gauntlet
```

## Binary

To easily use Gauntlet, we recommend to use the binary. To generate it, run:

```
yarn bundle
```

It will generate 2 binaries, for Linux and MacOS distributions. To use them, replace `yarn gauntlet` with `./bin/chainlink-starknet-<linux|macos>`

```bash
./bin/chainlink-starknet-macos

🧤  gauntlet 0.2.0
ℹ️   Available gauntlet commands:

example:
     example:deploy
     example:increase_balance
     example:inspect

account:
     account:deploy


ℹ️   Available global flags:

     --help, -h                             Display information about command usage
     --network                              The network to connect to
```

## Basic Setup

To deploy or query contracts you do not need any configuration. If you want to execute some contract method, you will need a wallet configured. The details should be added into a `.env` file in the root of the project.

```bash
## Public key of the account contract
ACCOUNT=0x...
## Private key of the wallet configured on the account contract
PRIVATE_KEY=0x...
```

In order to get this configuration, go to [how to setup an account](../../packages-ts/starknet-gauntlet-account/README.md#setup-an-account)

If you are interacting with a local network and do not want to use any wallet, execute every command with the flag `--noWallet`
