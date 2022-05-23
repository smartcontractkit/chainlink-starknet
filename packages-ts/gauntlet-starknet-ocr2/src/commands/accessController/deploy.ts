import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../lib/categories'
import { accessControllerContractLoader } from '../../lib/contracts'

type UserInput = {}

type ContractInput = {}

const makeUserInput = async (flags, args): Promise<UserInput> => {
  return {}
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [process.env.ACCOUNT]
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
