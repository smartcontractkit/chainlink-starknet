import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { ocr2ProxyLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

type UserInput = {}

type ContractInput = []

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return []
}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CATEGORIES.PROXY,
  category: CATEGORIES.PROXY,
  action: 'declare',
  ux: {
    description: `Declares an ${CATEGORIES.PROXY} contract`,
    examples: [`${CATEGORIES.PROXY}:declare --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: ocr2ProxyLoader,
}

export default makeExecuteCommand(commandConfig)
