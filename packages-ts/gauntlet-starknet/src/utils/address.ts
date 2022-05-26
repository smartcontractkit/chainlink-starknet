import { validateAndParseAddress } from 'starknet'

// TODO: This is inconsistent. come back here
export const isValidAddress = (address: string): boolean => {
  try {
    validateAndParseAddress(address)
    return true
  } catch (e) {}
  return false
}
