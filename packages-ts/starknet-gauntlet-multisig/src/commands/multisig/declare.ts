import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { contractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CATEGORIES.MULTISIG,
  category: CATEGORIES.MULTISIG,
  action: 'declare',
  ux: {
    description: 'Declares a SequencerUptimeFeed contract',
    examples: [`${CATEGORIES.MULTISIG}:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: contractLoader,
}

export default makeExecuteCommand(commandConfig)
