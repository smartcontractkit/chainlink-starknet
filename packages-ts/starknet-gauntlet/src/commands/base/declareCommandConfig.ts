import { ExecuteCommandConfig } from '.'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

export const declareCommandConfig = (
  contractId: string,
  category: string,
  contractLoader: any,
): ExecuteCommandConfig<UserInput, ContractInput> => ({
  contractId,
  category,
  action: 'declare',
  ux: {
    description: 'Declares the contract',
    examples: [`${contractId}:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: contractLoader,
})
