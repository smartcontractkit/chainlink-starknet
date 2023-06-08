import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST, aggregatorConsumerLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.AGGREGATOR_CONSUMER,
  category: CATEGORIES.OCR2,
  action: 'declare',
  suffixes: ['consumer'],
  ux: {
    description: `Declares an example Aggregator consumer`,
    examples: [`${CATEGORIES.OCR2}:consumer:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: aggregatorConsumerLoader,
}

export default makeExecuteCommand(commandConfig)
