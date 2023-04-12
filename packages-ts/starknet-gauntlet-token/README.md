# Gauntlet Starknet Commands for LINK Token

## Token

### Deploy

The contract is pre-configured to be the LINK token contract.

```bash
yarn gauntlet token:deploy --network=<NETWORK> --owner=<OWNER_ADDRESS>
```

IMPORTANT: For the token contract to be used in L1<>L2 bridging the `owner` must be the L2 Bridge address.

### Mint

```bash
yarn gauntlet token:mint --network=<NETWORK> --recipient=<RECPIENT_ACCOUNT> --amount=<AMOUNT> <TOKEN_CONTRACT_ADDRESS>
```

### Transfer

```bash
yarn gauntlet token:transfer --network=<NETWORK> --recipient=<RECPIENT_ACCOUNT> --amount=<AMOUNT> <TOKEN_CONTRACT_ADDRESS>
```

### Check balance

```bash
yarn gauntlet token:balance_of --network=<NETWORK> --address=<ACCOUNT_ADDRESS> <TOKEN_CONTRACT_ADDRESS>
```
