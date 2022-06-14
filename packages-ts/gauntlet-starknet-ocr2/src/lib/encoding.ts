import { number } from 'starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'

const CHUNK_SIZE = 31

export function bytesToFelts(data: Uint8Array): BN[] {
  let felts: BN[] = []

  // prefix with len
  let len = data.byteLength
  felts.push(number.toBN(len))

  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // cast to int
    felts.push(new BN(chunk, 'be'))
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
