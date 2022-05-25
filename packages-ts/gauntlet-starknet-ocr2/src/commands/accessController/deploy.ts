import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../lib/categories'
import { accessControllerContractLoader } from '../../lib/contracts'

type UserInput = {
  owner: string
}

type ContractInput = [owner: string]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  return {
    owner: flags.owner || env.account,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.owner]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.ACCESS_CONTROLLER,
    function: 'deploy',
    examples: [`${CATEGORIES.ACCESS_CONTROLLER}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: accessControllerContractLoader,
}

export default makeExecuteCommand(commandConfig)
