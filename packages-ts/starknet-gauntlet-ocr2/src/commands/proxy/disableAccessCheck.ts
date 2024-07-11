import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ProxyLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {}

type ContractInput = []

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {}
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.PROXY,
  category: CATEGORIES.PROXY,
  action: 'disable_access_check',
  ux: {
    description: 'Disable access check at aggregator proxy',
    examples: [`${CATEGORIES.PROXY}:disable_access_check --network=<NETWORK> <AGGREGATOR_PROXY>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: ocr2ProxyLoader,
}

export default makeExecuteCommand(commandConfig)
