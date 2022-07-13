# Starknet Gauntlet Commands for Starkgate Contracts

## ERC20

### Deploy the contract

```bash
yarn gauntlet starkgate_erc20:deploy --network=<NETWORK> --name=<NAME> --symbol=<SYMBOL> --decimals=<DECIMALS> "--minter=<MINTER_ADDRESS>"
# --minter is optional. If not provided, your default account contract will be used as minter
```

If you want to deploy a LINK contract, just include the `--link` flag:

```bash
yarn gauntlet starkgate_erc20:deploy --network=testnet --link
```

### Mint

```bash
yarn gauntlet starkgate_erc20:mint --network=<NETWORK> --recipient=<RECPIENT_ACCOUNT> --amount=<AMOUNT> <ERC20_CONTRACT_ADDRESS>
```

### Transfer

```bash
yarn gauntlet starkgate_erc20:transfer --network=<NETWORK> --recipient=<RECPIENT_ACCOUNT> --amount=<AMOUNT> <ERC20_CONTRACT_ADDRESS>
```
