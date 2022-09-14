import { constants, ec, encode, hash, number, uint256, stark, KeyPair } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
import { expect } from 'chai'
import { artifacts, network } from 'hardhat'

// This function adds the build info to the test network so that the network knows
// how to handle custom errors.  It is automatically done when testing
// against the default hardhat network.
export const addCompilationToNetwork = async (fullyQualifiedName: string) => {
  if (network.name !== 'hardhat') {
    // This is so that the network can know about custom errors.
    // Running against the provided hardhat node does this automatically.

    const buildInfo = await artifacts.getBuildInfo(fullyQualifiedName)
    if (!buildInfo) {
      throw Error('Cannot find build info')
    }
    const { solcVersion, input, output } = buildInfo
    console.log('Sending compilation result for StarkNetValidator test')
    await network.provider.request({
      method: 'hardhat_addCompilationResult',
      params: [solcVersion, input, output],
    })
    console.log('Successfully sent compilation result for StarkNetValidator test')
  }
}

export const expectInvokeError = async (invoke: Promise<any>, expected?: string) => {
  try {
    await invoke
  } catch (err: any) {
    expectInvokeErrorMsg(err?.message, expected)
    return // force
  }
  expect.fail("Unexpected! Invoke didn't error!?")
}

export const expectInvokeErrorMsg = (actual: string, expected?: string) => {
  // Match transaction error
  expect(actual).to.deep.contain('Transaction rejected. Error message:')
  // Match specific error
  if (expected) expectSpecificMsg(actual, expected)
}

export const expectCallError = async (call: Promise<any>, expected?: string) => {
  try {
    await call
  } catch (err: any) {
    expectCallErrorMsg(err?.message, expected)
    return // force
  }
  expect.fail("Unexpected! Call didn't error!?")
}

export const expectCallErrorMsg = (actual: string, expected?: string) => {
  // Match call error
  expect(actual).to.deep.contain('Could not perform call')
  // Match specific error
  if (expected) expectSpecificMsg(actual, expected)
}

export const expectSpecificMsg = (actual: string, expected: string) => {
  // Match specific error
  const matches = actual.match(/Error message: (.+?)\n/g)
  // Joint matches should include the expected, or fail
  if (matches && matches.length > 0) {
    expect(matches.join()).to.include(expected)
  } else expect.fail(`\nActual: ${actual}\n\nExpected: ${expected}`)
}

// Required to convert negative values into [0, PRIME) range
export const toFelt = (int: number | BigNumberish): BigNumberish => {
  const prime = number.toBN(encode.addHexPrefix(constants.FIELD_PRIME))
  return number.toBN(int).umod(prime)
}

// NOTICE: Leading zeros are trimmed for an encoded felt (number).
//   To decode, the raw felt needs to be start padded up to max felt size (252 bits or < 32 bytes).
export const hexPadStart = (data: number | bigint, len: number) => {
  return `0x${data.toString(16).padStart(len, '0')}`
}
