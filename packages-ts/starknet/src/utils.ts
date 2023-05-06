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
  // The error message is displayed as a felt hex string, so we need to convert the text.
  // ref: https://github.com/starkware-libs/cairo-lang/blob/c954f154bbab04c3fb27f7598b015a9475fc628e/src/starkware/starknet/business_logic/execution/execute_entry_point.py#L223
  const expectedHex = '0x' + Buffer.from(expected, 'utf8').toString('hex')
  const errorMessage = `Execution was reverted; failure reason: [${expectedHex}]`
  if (!actual.includes(errorMessage)) {
    expect.fail(`\nActual: ${actual}\n\nExpected:\n\tFelt hex: ${expectedHex}\n\tText: ${expected}`)
  }
}

// Starknet v0.11.0 and higher only allow declaring a class once:
// https://github.com/starkware-libs/starknet-specs/pull/85
export const expectSuccessOrDeclared = async (declareContractPromise: Promise<any>) => {
  try {
    await declareContractPromise
  } catch (err: any) {
    if (/Class with hash 0x[0-9a-f]+ is already declared\./.test(err?.message)) {
      return // force
    }
    expect.fail(err)
  }
}
