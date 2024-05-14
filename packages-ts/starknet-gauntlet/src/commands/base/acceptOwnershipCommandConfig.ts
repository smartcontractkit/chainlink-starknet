import { ExecuteCommandConfig } from '.'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

export const acceptOwnershipCommandConfig = (
  contractId: string,
  category: string,
  contractLoader: any,
): ExecuteCommandConfig<UserInput, ContractInput> => ({
  contractId,
  category: contractId,
  action: 'accept_ownership',
  ux: {
    description: 'End two-step ownership transfer process by accepting ownership',
    examples: [`${contractId}:accept_ownership --network=<NETWORK> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: contractLoader,
})
