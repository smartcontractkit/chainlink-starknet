import { expect } from 'chai'

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
  // Joint matches should contain the expected, or fail
  if (matches && matches.length > 1) {
    expect(matches.join()).to.contain(expected)
  } else expect.fail(actual, expected, 'Specific error message not found.')
}

/**
 * Receives a hex address, converts it to bigint, converts it back to hex.
 * This is done to strip leading zeros.
 * @param address a hex string representation of an address
 * @returns an adapted hex string representation of the address
 */
function adaptAddress(address: string) {
  return '0x' + BigInt(address).toString(16)
}

/**
 * Expects address equality after adapting them.
 * @param actual
 * @param expected
 */
export function expectAddressEquality(actual: string, expected: string) {
  expect(adaptAddress(actual)).to.equal(adaptAddress(expected))
}
