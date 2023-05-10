import { num, constants, cairo } from 'starknet'
import BN from 'bn.js'
import { encoding } from '@chainlink/gauntlet-contracts-ocr2'

const CHUNK_SIZE = 31

function packUint8Array(data: Uint8Array | Buffer): bigint {
  let result: bigint = BigInt(0)
  for (let i = 0; i < data.length; i++) {
    result = (result << BigInt(8)) | BigInt(data[i])
  }
  return result
}

export function bytesToFelts(data: Uint8Array | Buffer): string[] {
  const felts: string[] = []

  // prefix with data length
  felts.push(cairo.felt(data.byteLength))

  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // cairo.felt() does not support packing a Uint8Array natively.
    const packedValue = packUint8Array(chunk)
    felts.push(cairo.felt(packedValue))
  }
  return felts
}

export function bytesToFeltsDeprecated(data: Uint8Array): string[] {
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

const MAX_LEN: bigint = (BigInt(1) << BigInt(54)) - BigInt(1)

export function feltsToBytes(felts: string[]): Buffer {
  const data: number[] = []

  if (!felts.length) {
    throw new Error('Felt string is empty')
  }

  const remainingLengthBigInt = BigInt(felts.shift())
  if (remainingLengthBigInt > MAX_LEN) {
    throw new Error('Length does not fit in 54 bits')
  }

  let remainingLength = Number(remainingLengthBigInt)
  for (let felt of felts) {
    if (remainingLength <= 0) {
      throw new Error(
        `Too many felts (${felts.length}) for length ${remainingLengthBigInt.toString()}`,
      )
    }
    const chunkSize = Math.min(CHUNK_SIZE, remainingLength)
    let packedValue: bigint = BigInt(felt)
    const unpackedValues: number[] = []
    for (let i = 0; i < chunkSize; i++) {
      unpackedValues.push(Number(packedValue & BigInt(0xff)))
      packedValue = packedValue >> BigInt(8)
    }
    unpackedValues.reverse()
    data.push(...unpackedValues)
    remainingLength -= chunkSize
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
