import { expect } from 'chai'

export const assertErrorMsg = (full: string, expected: string) => {
  expect(full).to.deep.contain('Transaction rejected. Error message:')
  const match = /Error message: (.+?)\n/.exec(full)
  if (match && match.length > 1) {
    expect(match[1]).to.equal(expected)
    return
  }
  expect.fail('No expected error found: ' + expected)
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
