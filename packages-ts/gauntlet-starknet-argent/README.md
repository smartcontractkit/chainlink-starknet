# Gauntlet Starknet Commands for Argent Contracts

## Account

Deploy the contract:

```
yarn gauntlet argent_account:deploy --network=<NETWORK>
```

Note the contract address. The contract is not configured yet. A signer needs to be specified in it:

Initialize the contract:

```bash
yarn gauntlet argent_account:initialize --network=<NETWORK> <CONTRACT_ADDRESS>
# OR If you already have a private key
yarn gauntlet argent_account:initialize --network=<NETWORK> --publicKey=<PUBLIC_KEY> <CONTRACT_ADDRESS>
```

If no public key is provided, the command will generate a new Keypair and will give the details during the execution.

You need to pay some fee to call initialize, but as this could be the first account wallet you are deploying, use the `--noWallet` option to bypass the fee. This will be soon deprecated

At the end of the process, you will want to include the account contract and the private key to your `.env` configuration file.

```bash
# .env
PRIVATE_KEY=0x...
ACCOUNT=0x...
```
