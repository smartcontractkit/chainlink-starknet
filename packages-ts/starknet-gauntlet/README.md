# Starknet Gauntlet

## Overview

In the following guide, we'll provide a walk through on how to setup your environment to run Gauntlet.

## Prerequisites

First, you'll need to install the following software on your machine (please refer to the `.tool-versions` file in the repo's root directory for more info on the specific versions you should install):

- [Node.js](https://nodejs.org/en/download/package-manager) and [yarn](https://classic.yarnpkg.com/lang/en/docs/install)
- [Scarb](https://docs.swmansion.com/scarb/download.html)

## Getting Started

Once the necessary software has been installed, navigate to the root directory of the repo. From there, we can use the following command to build Gauntlet and compile both the Cairo and Solidity Starknet smart contracts:

```sh
make build-ts
```

Once everything has been built, you should be able to run Gauntlet:

```sh
yarn gauntlet
```

This should output a list of commands that come with Gauntlet:

```sh
yarn run v1.22.21
$ node ./packages-ts/starknet-gauntlet-cli/dist/index.js
.env not found
🧤  gauntlet 0.3.1
.env not found
ℹ️   Available gauntlet commands:

access_controller:
   access_controller:declare
   access_controller:deploy
   access_controller:upgrade
   access_controller:transfer_ownership
   access_controller:accept_ownership
   access_controller:declare:multisig
   access_controller:deploy:multisig

ocr2:
   ocr2:deploy
...
```

## Selecting an RPC URL

Before executing a Gauntlet command, you'll need a Starknet RPC v7 URL. You can find a list of common ones [here](https://www.starknetjs.com/docs/next/guides/connect_network/). Take note of one of these URLs and save it for later - we'll need it when we move onto configuring environment variables for Gauntlet.

## Setting up a Wallet

Gauntlet commands that send transactions to the chain require a funded Starknet wallet to cover transaction fees. If you're using testnet or mainnet, you can create a Starknet wallet by following the guide in the official documentation [here](https://docs.starknet.io/documentation/quick_start/set_up_an_account/). If you're using a docker container hosting a local Starknet node (i.e. see `./scripts/devnet.sh`), then you can use one of the predeployed accounts from the container logs.

## Funding your Wallet

If you're using testnet, you can fund your wallet using a Starknet faucet - a popular option is [this](https://starknet-faucet.vercel.app/) one. You can also reach out on Slack (#team-blockchain-integrations) to have the team send some funds directly to your account.

If you're using mainnet, please reach out on Slack for help funding your account.

If you're using a docker container hosting a local Starknet node (i.e. see `./scripts/devnet.sh`), all the predeployed accounts are funded by default - you can find the account details in the container logs):

```
| Account address |  0x64b48806902a367c8598f4f95c305e8c1a1acba5f082d294a43793113115691
| Private key     |  0x71d7bb07b9a64f6f78ac4c816aff4da9
| Public key      |  0x39d9e6ce352ad4530a0ef5d5a18fd3303c3606a7fa6ac5b620020ad681cc33b
```

## Populating Environment Variables

Now that we have an RPC v7 URL and a funded Starknet account, we'll need to provide them to Gauntlet via environment variables. Let's create a `.env` file in the repo's root directory:

```sh
touch .env
```

Once the file has been created, let's populate it with the values from the previous steps:

```env
# .env
NODE_URL=<rpc_url>
ACCOUNT=<account>
PRIVATE_KEY=<private_key>
```

Once these have been provided, we can start running commands. For example, to deploy the LINK token contract, we can use the following command:

```sh
yarn gauntlet token:deploy --link
```

This should produce the following logs:

```sh
yarn run v1.22.21
$ node ./packages-ts/starknet-gauntlet-cli/dist/index.js token:deploy --link
🧤  gauntlet 0.3.1
ℹ️   About to deploy the LINK Token Contract with the following details:
    0x64b48806902a367c8598f4f95c305e8c1a1acba5f082d294a43793113115691,0x64b48806902a367c8598f4f95c305e8c1a1acba5f082d294a43793113115691

ℹ️   Deploying contract token
🤔  Continue? (Y / N)Y
⏳  Sending transaction...
⏳  Waiting for tx confirmation at 0x15df49a78815ae44743f8a99ca7f3f9a7529594a76dc46853787b55ee4236b7...
✅  Contract deployed on 0x15df49a78815ae44743f8a99ca7f3f9a7529594a76dc46853787b55ee4236b7 with address 0x653d0d4c6969233b0f03095b7995793dbe7c7a7660e7cf426ecfc51aa42209f
ℹ️   If using RDD, change the RDD ID with the new contract address: 0x653d0d4c6969233b0f03095b7995793dbe7c7a7660e7cf426ecfc51aa42209f
ℹ️   Execution finished at transaction: 0x15df49a78815ae44743f8a99ca7f3f9a7529594a76dc46853787b55ee4236b7
✨  Done in 34.32s.
```

At this point, your environment should be fully configured to run Gauntlet commands.
