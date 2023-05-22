import { encoding } from '@chainlink/gauntlet-contracts-ocr2'
import { feltsToBytes } from '@chainlink/starknet-gauntlet'

export const decodeOffchainConfigFromEventData = (data: string[]): encoding.OffchainConfig => {
  // The ConfigSet event is defined as:
  // fn ConfigSet(
  //   previous_config_block_number: u64,
  //   latest_config_digest: felt252,
  //   config_count: u64,
  //   oracles: Array<OracleConfig>,
  //   f: u8,
  //   onchain_config: Array<felt252>,
  //   offchain_config_version: u64,
  //   offchain_config: Array<felt252>,
  //)
  const oraclesLenIndex = 3
  const oraclesLen = Number(BigInt(data[oraclesLenIndex]))
  const oracleStructSize = 2
  const fIndex = oraclesLenIndex + oraclesLen * oracleStructSize + 1
  const onchainConfigLenIndex = fIndex + 1
  const onchainConfigLen = Number(BigInt(data[onchainConfigLenIndex]))
  const offchainConfigVersionIndex = onchainConfigLenIndex + onchainConfigLen + 1
  const offchainConfigArrayLenIndex = offchainConfigVersionIndex + 1
  const offchainConfigStartIndex = offchainConfigArrayLenIndex + 1

  const offchainConfigFelts = data.slice(offchainConfigStartIndex)
  return encoding.deserializeConfig(feltsToBytes(offchainConfigFelts))
}
