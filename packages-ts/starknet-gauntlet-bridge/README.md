# Gauntlet Starknet Commands for LINK Token

## Bridge

All register/remove admin commands follow the same format

```bash
yarn gauntlet bridge:<REGISTER_OR_REMOVE_COMMAND> --account=<ADMIN_TO_REGISTER_OR_REMOVE> <CONTRACT>
```

Example:
```bash
yarn gauntlet bridge:register_governance_admin --account=<ADMIN_TO_REGISTER> <L2_BRIDGE_CONTRACT>
```



