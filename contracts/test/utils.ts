import { expect } from 'chai'

export const expectInvokeError = (full: string, expected: string) => {
  // Match transaction error
  expect(full).to.deep.contain('Transaction rejected. Error message:')
  // Match specific error
  const match = /Error message: (.+?)\n/.exec(full)
  if (match && match.length > 1) expect(match[1]).to.equal(expected)
  else expect.fail(`No expected error found: ${expected} \nFull error message: ${full}`)
}

export const expectCallError = (full: string, expected: string) => {
  // Match call error
  expect(full).to.deep.contain('Could not perform call')
  // Match specific error
  const match = /Error message: (.+?)\n/.exec(full)
  if (match && match.length > 1) expect(match[1]).to.equal(expected)
  else expect.fail(`No expected error found: ${expected} \nFull error message: ${full}`)
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
