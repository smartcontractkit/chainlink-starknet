import { cairo } from 'starknet'

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
