import { bytesToFelts, bytesToFeltsDeprecated, feltsToBytes } from '../../src/lib/encoding'

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
