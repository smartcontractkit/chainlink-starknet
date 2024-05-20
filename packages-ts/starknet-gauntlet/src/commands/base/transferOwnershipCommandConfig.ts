import { isValidAddress } from '../../utils'
import { ExecuteCommandConfig } from '.'

type ContractInput = [newOwner: string]

export interface TransferOwnershipUserInput {
  newOwner: string
}

const validateNewOwner = async (input) => {
  if (isValidAddress(input.newOwner)) {
    return true
  }
  throw new Error(`Invalid New Owner Address: ${input.newOwner}`)
}

const makeUserInput = async (flags): Promise<TransferOwnershipUserInput> => {
  if (flags.input) return flags.input as TransferOwnershipUserInput
  return {
    newOwner: flags.newOwner,
  }
}

const makeContractInput = async (input: TransferOwnershipUserInput): Promise<ContractInput> => {
  return [input.newOwner]
}

export const transferOwnershipCommandConfig = (
  contractId: string,
  category: string,
  contractLoader: any,
): ExecuteCommandConfig<TransferOwnershipUserInput, ContractInput> => ({
  contractId,
  category: contractId,
  action: 'transfer_ownership',
  ux: {
    description: 'Begin two-step ownership transfer process by proposing pending owner',
    examples: [
      `${contractId}:transfer_ownership --network=<NETWORK> --newOwner=<NEW_OWNER_ADDRESS> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateNewOwner],
  loadContract: contractLoader,
})
