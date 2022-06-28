import { number } from 'starknet'
import BN from 'bn.js'
import { encoding } from '@chainlink/gauntlet-contracts-ocr2'

const CHUNK_SIZE = 31

export function bytesToFelts(data: Uint8Array): string[] {
  let felts: string[] = []

  // prefix with len
  let len = data.byteLength
  felts.push(number.toBN(len).toString())

  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // cast to int
    let bn = new BN(chunk, 'be')
    felts.push(bn.toString())
  }
  return felts
}

export function feltsToBytes(felts: BN[]): Buffer {
  let data = []

  // TODO: validate len > 1

  // TODO: validate it fits into 54 bits
  let length = felts.shift()?.toNumber()!

  for (const felt of felts) {
    let chunk = felt.toArray('be', Math.min(CHUNK_SIZE, length))
    data.push(...chunk)

    length -= chunk.length
  }

  return Buffer.from(data)
}

export const decodeOffchainConfigFromEventData = (data: string[]): encoding.OffchainConfig => {
  const oraclesLen = number.toBN(data[3]).toNumber()
  /** SetConfig event data has the following info:
    previous_config_block_number=prev_block_num,
    latest_config_digest=digest,
    config_count=config_count,
    oracles_len=oracles_len, // It includes both signer and transmitter addresses
    oracles=oracles,
    f=f,
    onchain_config=onchain_config,
    offchain_config_version=offchain_config_version,
    offchain_config_len=offchain_config_len,
   */
  const offchainConfigFelts = data.slice(3 + oraclesLen * 2 + 5)
  return encoding.deserializeConfig(feltsToBytes(offchainConfigFelts.map((f) => number.toBN(f))))
}
