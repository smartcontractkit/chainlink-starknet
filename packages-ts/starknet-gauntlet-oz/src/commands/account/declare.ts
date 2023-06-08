import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { accountContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CATEGORIES.ACCOUNT,
  category: CATEGORIES.ACCOUNT,
  action: 'declare',
  ux: {
    description: `Declares an ${CATEGORIES.ACCOUNT} contract`,
    examples: [`${CATEGORIES.ACCOUNT}:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: accountContractLoader,
}

export default makeExecuteCommand(commandConfig)
