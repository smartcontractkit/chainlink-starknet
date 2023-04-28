import { constants, encode, num } from 'starknet'
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
    console.log('Sending compilation result for StarknetValidator test')
    await network.provider.request({
      method: 'hardhat_addCompilationResult',
      params: [solcVersion, input, output],
    })
    console.log('Successfully sent compilation result for StarknetValidator test')
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
  expect(actual).to.deep.contain('TRANSACTION_FAILED')
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
