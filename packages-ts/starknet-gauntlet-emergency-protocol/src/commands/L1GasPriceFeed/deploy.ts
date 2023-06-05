import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { L1MockAggregatorLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {}

type ContractInput = []

const makeUserInput = async (flags, args, env): Promise<UserInput> => ({})

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_MOCK_AGGREGATOR,
  category: CATEGORIES.L1_MOCK_AGGREGATOR,
  action: 'deploy',
  ux: {
    description: 'Deploy an MockAggregator',
    examples: [`${CATEGORIES.L1_MOCK_AGGREGATOR}:deploy --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: L1MockAggregatorLoader,
}

export default makeEVMExecuteCommand(commandConfig)
