import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../lib/categories'
import { accessControllerContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  owner: string
}

type ContractInput = [owner: string]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    owner: flags.owner || env.account,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.owner]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCESS_CONTROLLER,
  category: CATEGORIES.ACCESS_CONTROLLER,
  action: 'deploy',
  ux: {
    description: 'Deploys an Access Controller Contract',
    examples: [`${CATEGORIES.ACCESS_CONTROLLER}:deploy --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: accessControllerContractLoader,
}

export default makeExecuteCommand(commandConfig)
