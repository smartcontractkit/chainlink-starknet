# Gauntlet Starknet Commands for a Sample Contract

This package contains some commands to serve as an example on how to use Gauntlet with Starknet. The contract can be found on the [Cairo Docs](https://www.cairo-lang.org/docs/hello_starknet/intro.html#your-first-contract)

## Commands

- Deploy

```bash
yarn gauntlet example:deploy --network=<NETWORK>
```

This will result in a new contract address

- Increasing balance
  
Will increase the current balance with the amount specified with the `--balance` flag
```bash
yarn gauntlet example:increase_balance --network=<NETWORK> --balance=<NUMBER> (--noWallet) <CONTRACT_ADDRESS>
```

- Inspect the contract
  
Will prompt the current balance of the contract
```bash
yarn gauntlet example:inspect --network=<NETWORK> <CONTRACT_ADDRESS>
```

