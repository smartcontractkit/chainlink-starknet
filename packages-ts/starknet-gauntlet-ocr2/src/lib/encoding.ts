import { num, constants, encode } from 'starknet'
import BN from 'bn.js'
import { encoding } from '@chainlink/gauntlet-contracts-ocr2'

const CHUNK_SIZE = 31

export function bytesToFelts(data: Uint8Array): string[] {
  let felts: string[] = []

  // prefix with len
  let len = data.byteLength
  felts.push(num.toBigInt(len).toString())

  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // cast to int
    felts.push(new BN(chunk, 'be').toString())
  }
  return felts
}

export function feltsToBytes(felts: string[]): Buffer {
  console.log(felts)
  let data = []

  // TODO: validate len > 1

  // TODO: validate it fits into 54 bits
  let length = Number(BigInt(felts.shift()))

  for (const felt of felts) {
    let size = Math.min(CHUNK_SIZE, length)
    let chunk = Uint8Array.from(Buffer.from(felt, 'hex'))
    let filler = chunk.length - size
    // pad to size
    while (filler > 0) {
      data.push(0)
    }
    data.push(...chunk)

    length -= chunk.length
  }

  return Buffer.from(data)
}

export const decodeOffchainConfigFromEventData = (data: string[]): encoding.OffchainConfig => {
  const oraclesLen = Number(num.toBigInt(data[3]))
  /** SetConfig event data has the following info:
    0 : previous_config_block_number=prev_block_num,
    1 : latest_config_digest=digest,
    2 : config_count=config_count,
    3 : oracles_len=oracles_len, // It includes both signer and transmitter addresses
    3 + 2X : oracles=oracles,
    3 + 2X + 1 : f=f,
    3 + 2X + 2 + 3 : onchain_config_len=OnchainConfig.SIZE = 3
    onchain_config=onchain_config,
    3 + 2X + 2 + 3 + 1 : offchain_config_version=offchain_config_version,
    3 + 2X + 2 + 3 + 2 : offchain_config_len=offchain_config_len,
    3 + 2X + 2 + 3 + 3 : offchain_config=offchain_config
   */
  const offchainConfigFelts = data.slice(3 + oraclesLen * 2 + 8)
  return encoding.deserializeConfig(feltsToBytes(offchainConfigFelts))
}
