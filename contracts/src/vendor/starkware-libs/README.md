# starkware-libs vendor contracts

Here we duplicate the `starkware-libs/starkgate-contracts` project as we couldn't find a way to properly import it via NPM (only a few .cairo contracts packaged), PyPI (N/A), or a submodule (complex build).

- `starkware-libs/starkgate-contracts` - duplicate of the original repo at [c08863a](https://github.com/starkware-libs/starkgate-contracts/commit/c08863a1f08226c09f1d0748124192e848d73db9) (includes only the files we use)
- `starkware-libs/starkgate-contracts-solidity-v0.8` - fork of the original repo at [c08863a](https://github.com/starkware-libs/starkgate-contracts/commit/c08863a1f08226c09f1d0748124192e848d73db9) which loosens the `pragma` declaration for a few interfaces to support v0.8 (includes only the files we use)
