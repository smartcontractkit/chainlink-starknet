import { isValidAddress } from '@chainlink/starknet-gauntlet'

export const validateClassHash = async (input) => {
  if (isValidAddress(input.classHash) || input.classHash === undefined) {
    return true
  }
  throw new Error(`Invalid Class Hash: ${input.classHash}`)
}
