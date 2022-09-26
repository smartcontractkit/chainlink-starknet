import { validateAndParseAddress } from 'starknet'

// TODO: This is inconsistent. come back here
export const isValidAddress = (address: string): boolean => {
  try {
    validateAndParseAddress(address)
    return !!address // check value is not falsy (undefined, "", etc)
  } catch (e) {}
  return false
}
