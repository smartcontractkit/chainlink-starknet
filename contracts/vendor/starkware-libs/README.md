# starkware-libs vendor contracts

- `starkware-libs/starkgate-contracts` - fork of the original repo at [c08863a](https://github.com/starkware-libs/starkgate-contracts/commit/c08863a1f08226c09f1d0748124192e848d73db9) includes only `std_contracts/ERC20/permitted.cairo`
- `starkware-libs/cairo-lang` fork of the original repo at [v0.11.0.2](https://github.com/starkware-libs/cairo-lang/tree/v0.11.0.2/src/starkware/starknet) which loosens the `pragma` declaration for a few interfaces to support v0.8 (includes only the files we use)