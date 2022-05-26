import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  Validation,
} from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../lib/categories'
import { accountContractLoader } from '../../lib/contracts'

type UserInput = {}

type ContractInput = []

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  return []
}

const validate: Validation<UserInput> = async (input) => {
  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to deploy an Argent Account Contract`)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.ACCOUNT,
    function: 'deploy',
    examples: [`${CATEGORIES.ACCOUNT}:deploy --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  loadContract: accountContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
