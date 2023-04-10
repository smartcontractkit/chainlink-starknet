# Gauntlet Starknet Commands for Starkgate Contracts

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

## Token Bridge

### Deploy

Token bridge deployment requires communicating to both L1 and L2 networks. Before executing the command make sure the environment contains the keys relevant to the network you execute the command for (the keys can be stored in the corresponding .env.\<network\> file).

#### Deploy L1 bridge

```bash
yarn gauntlet l1_bridge:deploy --network=<L1_NETWORK>
```

#### Deploy L1 proxy and initialize L1 bridge

```bash
yarn gauntlet l1_bridge:deploy:proxy --network=<L1_NETWORK> --bridge=<L1_BRIDGE_ADDRESS> --token=<L1_TOKEN_ADDRESS> --core=<STARKNET_CORE_CONTRACT_ADDRESS>
```

#### Deploy L2 bridge

```bash
yarn gauntlet l2_bridge:deploy --network=<L2_NETWORK> --governor=<GOVERNOR_ADDRESS>
# --governor is optional. If not provided, your default account will be used as governor
```

#### Set L2 Token

WARNING: the command can only be executed once.

```bash
yarn gauntlet l2_bridge:set_l2_token --network=<L2_NETWORK> --address=<L2_TOKEN_ADDRESS> <L2_BRIDGE_ADDRESS>
```

#### Set L1 bridge

WARNING: The command can only be executed once.

```bash
yarn gauntlet l2_bridge:set_l1_bridge --network=<L2_NETWORK> --address=<L1_BRIDGE_PROXY_ADDRESS> <L2_BRIDGE_ADDRESS>
```

#### Set L2 bridge

WARNING: The command can only be executed once.

```bash
yarn gauntlet l1_bridge:set_l2_bridge --network=<L1_NETWORK> --address=<L2_BRIDGE_ADDRESS> <L1_BRIDGE_PROXY_ADDRESS>
```

#### Configure L1 Bridge

Maximum total balance deposited on the bridge:

```bash
yarn gauntlet l1_bridge:set_max_total_balance --network=<L1_NETWORK> --amount=<AMOUNT_IN_LINK> <L1_BRIDGE_PROXY_ADDRESS>
```

Maximum deposit:

```bash
yarn gauntlet l1_bridge:set_max_deposit --network=<L1_NETWORK> --amount=<AMOUNT_IN_LINK> <L1_BRIDGE_PROXY_ADDRESS>
```

Additionally, L1 token contract must be configured to allow L1 token bridge to spend (`token:approve` command).

### Deposit

```bash
yarn gauntlet l1_bridge:deposit --network=<L1_NETWORK> --amount=<AMOUNT_IN_LINK> --recipient=<L2_RECIPIENT_ADDRESS> <L1_BRIDGE_PROXY_ADDRESS>
```

### Withdraw

```bash
yarn gauntlet l2_bridge:initiate_withdraw --network=<L2_NETWORK> --recipient=<L1_RECIPIENT_ADDRESS> --amount=<AMOUNT_IN_LINK> <L2_BRIDGE_ADDRESS>
```

```bash
yarn gauntlet l1_bridge:withdraw --network=<L1_NETWORK> --recipient=<L1_RECIPIENT_ADDRESS> --amount=<AMOUNT_IN_LINK> <L1_BRIDGE_PROXY_ADDRESS>
```
