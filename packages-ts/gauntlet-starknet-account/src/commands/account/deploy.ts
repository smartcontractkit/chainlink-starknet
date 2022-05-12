import { ExecuteCommandConfig, makeExecuteCommand, Validation } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../lib/categories'
import { accountContract } from '../../lib/contracts'

type UserInput = {
  address?: string
}

type ContractInput = {}

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  // If address is provided, deployContract should use that as adressSalt
  return {}
}

const validate: Validation<UserInput> = async (input) => {
  return true
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.ACCOUNT,
    function: 'deploy',
    examples: [`${CATEGORIES.ACCOUNT}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  contract: accountContract,
}

export default makeExecuteCommand(commandConfig)
