import { isValidAddress } from '../../utils'
import { ExecuteCommandConfig } from '.'

type ContractInput = [classHash: string]

export interface UserInput {
  classHash: string
}

const validateClassHash = async (input) => {
  if (isValidAddress(input.classHash)) {
    return true
  }
  throw new Error(`Invalid Class Hash: ${input.classHash}`)
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    classHash: flags.classHash,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.classHash]
}

export const upgradeCommandConfig = (
  contractId: string,
  category: string,
  contractLoader: any,
): ExecuteCommandConfig<UserInput, ContractInput> => ({
  contractId,
  category: contractId,
  action: 'upgrade',
  ux: {
    description: 'Upgrades contract to new class hash',
    examples: [
      `${contractId}:upgrade --network=<NETWORK> --classHash=<CLASS_HASH> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateClassHash],
  loadContract: contractLoader,
})
