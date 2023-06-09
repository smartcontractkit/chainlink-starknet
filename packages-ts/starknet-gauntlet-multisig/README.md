# Starknet Gauntlet Commands for Multisig Contract

## Multisig

### Deploy the contract

```bash
yarn gauntlet multisig:deploy --network=<NETWORK> --threshold=<APPROVALS_NEEDED> --signers=<[SIGNERS_LIST]>
```

Note that threshold must be equal or higher than the amount of signers

Examples:

```bash
yarn gauntlet multisig:deploy --network=local --threshold=2 --signers="[0x26e10005e67c478b373658755749a60f2f31bc955a6a2311eb456b20b8913e9, 0x56bfff7e282d1e023c6268e72dba551a22c1bf816a30334ae43b5c491c99bb8]"
# Or
yarn gauntlet multisig:deploy --network=local --input='{"threshold":2, "signers": ["0x26e10005e67c478b373658755749a60f2f31bc955a6a2311eb456b20b8913e9", "0x56bfff7e282d1e023c6268e72dba551a22c1bf816a30334ae43b5c491c99bb8"]}'
```

### Set signers

Signers can only be updated through a transaction executed from the multisig itself

```bash
yarn gauntlet multisig:set_signers:multisig --network=<NETWORK> --signers=<[SIGNERS_LIST]> <MULTISIG_CONTRACT_ADDRESS>
```

### Set threshold

Threshold can only be updated through a transaction executed from the multisig itself.

```bash
yarn gauntlet multisig:set_thresold:multisig --network=<NETWORK> --threshold=<APPROVALS_NEEDED> <MULTISIG_CONTRACT_ADDRESS>
```

### Upgrade

To upgrade the multisig, you will need to create a proposal calling the `upgrade` function on the multisig contract. Please read the instructions below for how to create a proposal

## Wrapping Gauntlet commands

A [wrap function](./src/wrapper/index.ts#L30) is exposed from this package. It allows to wrap any Gauntlet command and make its functionality available to be executed from a multisig wallet. The process is the same for every command:

1. Create proposal

```bash
yarn gauntlet <CATEGORY>:<FUNCTION>:multisig --network=<NETWORK> (...<INPUT NEEDED FOR COMMAND>) <CONTRACT_ADDRESS>
```

This will create a proposal in the multisig contract, that needs to be approved and executed.
The proposal index will be prompted.

2. Approve proposal

`T` (threshold) signers need to run this command

```bash
yarn gauntlet <CATEGORY>:<FUNCTION>:multisig --network=<NETWORK> (...<INPUT NEEDED FOR COMMAND>) --multisigProposal=<PROPOSAL_ID> <CONTRACT_ADDRESS>
```

3. Execute proposal

Once approvals have reached the threshold, the proposal can be executed.

```bash
yarn gauntlet <CATEGORY>:<FUNCTION>:multisig --network=<NETWORK> (...<INPUT NEEDED FOR COMMAND>) --multisigProposal=<PROPOSAL_ID> <CONTRACT_ADDRESS>
```

If you are running these commands from the `starknet-gauntlet-cli`, you'll want to add to your environment the Multisig address:

```bash
# .env
...
MULTISIG=<MULTISIG_CONTRACT_ADDRESS>
```
