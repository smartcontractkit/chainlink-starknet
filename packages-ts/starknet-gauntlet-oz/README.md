# Starknet Gauntlet Commands for the Open Zeppelin Contracts

## Account

### Deploy

```bash
yarn gauntlet account:deploy --network=<NETWORK>
```

This command will generate a new Keypair and will give the details during the execution. If you already have a Keypair and want to use it as signer on your Account contract, use:

```bash
yarn gauntlet account:deploy --network=<NETWORK> --publicKey=<PUBLIC_KEY>
```

After the execution is finished, you will want to include the account contract and the private key to your `.env` configuration file.

```bash
# .env
PRIVATE_KEY=0x...
ACCOUNT=0x...
```
