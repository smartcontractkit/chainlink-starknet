import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { L1MockAggregatorLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  answer: number
}

type ContractInput = [number]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  return {
    answer: flags.answer,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [Number(input.answer)]
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_MOCK_AGGREGATOR,
  category: CATEGORIES.L1_MOCK_AGGREGATOR,
  action: 'set_answer',
  internalFunction: 'setLatestAnswer',
  ux: {
    description: 'Set the static LatestAnswer',
    examples: [
      `${CATEGORIES.L1_MOCK_AGGREGATOR}:set_answer --network=<NETWORK> --answer=<ANSWER> <AGGREGATOR_CONTRACT>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: L1MockAggregatorLoader,
}

export default makeEVMExecuteCommand(commandConfig)
