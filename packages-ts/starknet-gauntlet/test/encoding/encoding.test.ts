import { num } from 'starknet'
import { bytesToFelts, feltsToBytes } from '../../src/encoding'

const CHUNK_SIZE = 31

export function bytesToFeltsDeprecated(data: Uint8Array): string[] {
  let felts: string[] = []
  // prefix with len
  let len = data.byteLength
  felts.push(num.toBigInt(len).toString())
  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // convert big int to int (decimal) string
    felts.push(bytesToBigInt(chunk).toString())
  }
  return felts
}

function bytesToBigInt(data: Uint8Array): bigint {
  // convert byte array to hexadecimal string (pad by 2 so each byte is represented by 2 hex characters)
  const feltHex = `0x${Array.from(data, (byte) => byte.toString(16).padStart(2, '0')).join('')}`
  // convert hexadecimal string to bigint
  return BigInt(feltHex)
}

function createUint8Array(len: number): Uint8Array {
  const ret: Uint8Array = new Uint8Array(len)
  for (let i = 1; i <= len; i++) {
    ret[i - 1] = i
  }
  return ret
}

describe('bytesToFelts', () => {
  it('matches the deprecated BN function', async () => {
    for (let testLength = 0; testLength < 256; testLength++) {
      const testArray = createUint8Array(testLength)
      expect(bytesToFelts(testArray)).toEqual(bytesToFeltsDeprecated(testArray))
    }
  })

  it('converts to felts and back successfully', async () => {
    for (let testLength = 0; testLength < 256; testLength++) {
      const testArray = createUint8Array(testLength)
      expect(new Uint8Array(feltsToBytes(bytesToFelts(testArray)))).toEqual(testArray)
    }
  })
})
