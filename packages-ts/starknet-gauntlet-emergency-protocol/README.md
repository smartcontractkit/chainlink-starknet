# Starknet Gauntlet Commands to manage the Starknet Emergency Protocol

This package contains the commands required to manage the contracts related to the Starknet Emergency Protocol

## Commands

### StarknetValidator

- Deploy

This deploys a new instance of the `StarknetValidator` contract on **L1**

`<STARKNET_MESSAGING>` is address of the official Starkware Industries deployed messaging contract (ex: 0xde29d060D45901Fb19ED6C6e959EB22d8626708e on goerli testnet)

`<CONFIG_AC>` is address of access controller which can modify the StarknetValidator config

`<GAS_PRICE_L1_FEED>` is adddress of fast gas data feed (ex: 0x169E633A2D1E6c10dD91238Ba11c4A708dfEF37C on mainnet)

`<SOURCE>` is the address of source aggregator. Source aggregator should be added to access control of starknet validator contract to be able to call
`validate` on the starknet validator

`<GAS_ESTIMATE>` is a number that represents l1 gas estimate of the cost of doing l2 work. As of the time of writing, we recommend a value of 17300 on mainnet (for the cairo-1 contract version SequencerUptimeFeed 1.0.0). On devnet or testnet you may choose 1 to save gas.

`<L2_FEED>` is the layer 2 feed

`<GAS_ADJUSTMENT>` is the percentage adjustment made to the gas cost. For example, a value of 110 would equate to 110% of the original gas cost (or equivalently, a 10% bump in price, or equivalently a 1.1 times the original cost). For simplicity sake recommend a value of 130 on mainnet, devnet, or testnet. If you set it below 100, you'll need to pass in the `--unsafe` flag as well.

`(Optional) --unsafe`: Flag passed in to specify a gas adjustment below 100

```bash
yarn gauntlet starknet_validator:deploy --starkNetMessaging=<STARKNET_MESSAGING> --configAC=<CONFIG_AC> --gasPriceL1Feed=<GAS_PRICE_L1_FEED> --source=<SOURCE_AGGREGATOR> --gasEstimate=<GAS_ESTIMATE> --l2Feed=<L2_FEED> --gasAdjustment=<GAS_ADJUSTMENT> --network=<NETWORK>
```

- Accept Ownership

Will accept ownership of the contract. This should be done after the current owner transfers ownership.

```bash
yarn gauntlet starknet_validator:accept_ownership --network=<NETWORK> <CONTRACT_ADDRESS>
```

- Transfers Ownership

Will transfer ownership to a new owner. The new owner must accept ownership to take control of the contract.

```bash
yarn gauntlet starknet_validator:transfer_ownership --to=<NEW_PROPOSED_OWNER> <CONTRACT_ADDRESS> --network=<NETWORK>
```

- Add Access

Allows an address to write to the validator

```bash
yarn gauntlet starknet_validator:add_access --address=<ADDRESS> --network=<NETWORK> <CONTRACT_ADDRESS>
```

- Validate

Calls validate on the validator contract

```bash
yarn gauntlet starknet_validator:validate --previousRoundId=<PREV_ROUND_ID> --previousAnswer=<PREV_ANSWER> --currentRoundId=<CURR_ROUND_ID> --currentAnswer=<CURR_ANSWER> <VALIDATOR_ADDRESS> --network=<NETWORK>

```

### Sequencer Uptime Feed

- Deploy

This deploys a new `sequencer_uptime_feed` contract to L2.

`<INITIAL_STATUS>` can be 0 or 1. 0 means that feed is healthy and up.

`--owner` flag can be omitted. In such a case, it will default to the account specified in .env

```bash
yarn gauntlet sequencer_uptime_feed:deploy --initialStatus=<INITIAL_STATUS> --owner=<OWNER> --network=<NETWORK>
```

- setL1Sender

This sets the L1 sender address. This is to control, which L1 address can write new statuses to the uptime feed.

--address is the L1 sender address, which should be the deployed StarknetValidator.sol contract

```bash
yarn gauntlet sequencer_uptime_feed:set_l1_sender --network=<NETWORK> --address=<ADDRESS>  <L2_FEED>
```

- upgrade

This upgrades the contract to point to a new class hash. 

```bash
yarn gauntlet SequencerUptimeFeed:upgrade --network=testnet --classHash=<CLASS_HASH>
```

- Inspect

Inspect the latest round data

```bash
yarn gauntlet sequencer_uptime_feed:inspect --network=<NETWORK> <CONTRACT_ADDRESS>
```
