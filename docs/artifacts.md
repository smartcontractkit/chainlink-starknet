# Artifacts

## A Note on Starknet Artifacts

In `starknet.js` v6.7.0, it is necessary to supply both the sierra and sierra casm compilation outputs to a declare transaction as shown [here](https://www.starknetjs.com/docs/next/guides/create_contract#declare-for-a-new-class).

The [`declare`](https://github.com/starknet-io/starknet.js/blob/a85d48ee73acb1365da6bef3f9d3a65153f9a422/src/account/default.ts#L393) method uses the [`artifact` field to compute the `class_hash`](https://github.com/starknet-io/starknet.js/blob/a85d48ee73acb1365da6bef3f9d3a65153f9a422/src/utils/contract.ts#L38), and the [`casm` field to compute the `compiled_class_hash`](https://github.com/starknet-io/starknet.js/blob/a85d48ee73acb1365da6bef3f9d3a65153f9a422/src/utils/contract.ts#L30).

For V2 and V3 declare transactions, both the `class_hash` and `compiled_class_hash` are required to construct the tx hash:

- [V3 declare transaction docs](https://docs.starknet.io/documentation/architecture_and_concepts/Network_Architecture/transactions/#v3_hash_calculation_2)
- [V2 declare transaction docs](https://docs.starknet.io/documentation/architecture_and_concepts/Network_Architecture/transactions/#v2_deprecated_hash_calculation)
  - Side note: even though V2 is deprecated, I'm still including it bc [`starknet.js` defaults to transaction version 2 for backwards compatibility in v6.7.0](https://github.com/starknet-io/starknet.js/blob/a85d48ee73acb1365da6bef3f9d3a65153f9a422/src/account/default.ts#L75)

If you attempt to pass only the casm artifact alone or the sierra artifact alone, this causes an error like the one below:

```sh
"Extract compiledClassHash failed, provide (CairoAssembly).casm file or compiledClassHash"
```
