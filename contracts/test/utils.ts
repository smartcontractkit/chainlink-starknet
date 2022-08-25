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
