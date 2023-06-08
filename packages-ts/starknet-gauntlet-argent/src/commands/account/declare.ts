import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST, accountContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCOUNT,
  category: CATEGORIES.ACCOUNT,
  action: 'declare',
  ux: {
    description: 'Declares a SequencerUptimeFeed contract',
    examples: [`${CATEGORIES.ACCOUNT}:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: accountContractLoader,
}

export default makeExecuteCommand(commandConfig)
