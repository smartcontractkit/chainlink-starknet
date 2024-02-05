import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { L1AccessControllerContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {}

type ContractInput = []

const makeUserInput = async (flags, args, env): Promise<UserInput> => ({})

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_ACCESS_CONTROLLER,
  category: CATEGORIES.L1_ACCESS_CONTROLLER,
  action: 'deploy',
  ux: {
    description: 'Deploy an AccessController',
    examples: [`${CATEGORIES.L1_ACCESS_CONTROLLER}:deploy --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: L1AccessControllerContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
