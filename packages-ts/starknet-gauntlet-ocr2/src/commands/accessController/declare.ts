import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { accessControllerContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CATEGORIES.ACCESS_CONTROLLER,
  category: CATEGORIES.ACCESS_CONTROLLER,
  action: 'declare',
  ux: {
    description: `Declares an ${CATEGORIES.ACCESS_CONTROLLER} contract`,
    examples: [`${CATEGORIES.ACCESS_CONTROLLER}:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: accessControllerContractLoader,
}

export default makeExecuteCommand(commandConfig)
