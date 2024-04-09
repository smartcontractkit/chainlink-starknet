import { encoding } from '@chainlink/gauntlet-contracts-ocr2'
import { feltsToBytes } from '@chainlink/starknet-gauntlet'
import { BigNumberish } from 'starknet'

export const decodeOffchainConfigFromEventData = (
  data: BigNumberish[],
): encoding.OffchainConfig => {
  return encoding.deserializeConfig(feltsToBytes(data))
}
