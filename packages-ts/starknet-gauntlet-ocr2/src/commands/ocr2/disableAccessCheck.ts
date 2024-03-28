import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader, CONTRACT_LIST } from '../../lib/contracts'

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
  contractId: CONTRACT_LIST.OCR2,
  category: CATEGORIES.OCR2,
  action: 'disable_access_check',
  ux: {
    description: 'Disable access check at aggregator',
    examples: [`${CATEGORIES.OCR2}:disable_access_check --network=<NETWORK> <AGGREGATOR>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
