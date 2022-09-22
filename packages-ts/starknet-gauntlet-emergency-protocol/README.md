# Starknet Gauntlet Commands to manage the Starknet Emergency Protocol

This package contains the commands required to manage the contracts related to the Starknet Emergency Protocol

## Commands

### StarknetValidator

- Deploy

This deploys a new instance of the `StarknetValidator` contract on **L1**

```bash
yarn gauntlet StarknetValidator:deploy --starkNetMessaging=0xde29d060D45901Fb19ED6C6e959EB22d8626708e --configAC=0x42f4802128C56740D77824046bb13E6a38874331 --gasPriceL1Feed=0xdcb95Cd00d32d02b5689CE020Ed67f4f91ee5942 --source=0x42f4802128C56740D77824046bb13E6a38874331 --gasEstimate=0 --l2Feed=0x0646bbfcaab5ead1f025af1e339cb0f2d63b070b1264675da9a70a9a5efd054f --network=
```

- Accept Ownership

Will accept ownership of the contract. This should be done after the current owner transfers ownership.

```bash
yarn gauntlet StarknetValidator:accept_ownership 0xAD6F411BF8559002CC9800A2E9aA87A0ff1b464e --network=<NETWORK>
```

- Transfers Ownership

Will transfer ownership to a new owner. The new owner must accept ownership to take control of the contract.

```bash
yarn gauntlet StarknetValidator:transfer_ownership --to=0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4 0xAD6F411BF8559002CC9800A2E9aA87A0ff1b464e --network=<NETWORK>
```

- Add Access

Allows an address to write to the validator

```bash
yarn gauntlet StarknetValidator:add_access --user=0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4 0x6B5b7121C4F4B186e8C018a65CF379260B0Dba04 --network=<NETWORK>
```

### Sequencer Uptime Feed

- Deploy

This deploys a new `sequencer_uptime_feed` contract to L2.

```bash
yarn gauntlet sequencer_uptime_feed:deploy --initialStatus=false --network=<NETWORK>
```

- setL1Sender

This sets the L1 sender address. This is to control, which L1 address can write new statuses to the uptime feed.

```bash
yarn gauntlet sequencer_uptime_feed:set_l1_sender --network=<NETWORK> --address=0x31982C9e5edd99bb923a948252167ea4BbC38AC1 0x0646bbfcaab5ead1f025af1e339cb0f2d63b070b1264675da9a70a9a5efd054f
```
