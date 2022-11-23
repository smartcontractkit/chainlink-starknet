import {
  EVMExecuteCommandConfig,
  EVMExecutionContext,
  makeEVMExecuteCommand,
} from '@chainlink/evm-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { l1BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {}
type ContractInput = any[]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {}
}

const makeContractInput = async (
  input: UserInput,
  context: EVMExecutionContext,
): Promise<ContractInput> => {
  return []
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_BRIDGE,
  category: CATEGORIES.L1_BRIDGE,
  action: 'deploy',
  ux: {
    description: 'Deploys an L1 token bridge',
    examples: [`${CATEGORIES.L1_BRIDGE}:deploy --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: l1BridgeContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
