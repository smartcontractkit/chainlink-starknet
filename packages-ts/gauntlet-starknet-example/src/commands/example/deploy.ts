import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader } from '../../lib/contracts'

type UserInput = {}

type ContractInput = {}

const makeUserInput = async (flags, args): Promise<UserInput> => ({})

const makeContractInput = async (input: UserInput): Promise<ContractInput> => ({})

// This is a custom beforeExecute hook executed right before the command action is executed
const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info('About to deploy a Sample Contract')
  await deps.prompt('Continue?')
}
// This is a custom afterExecute hook executed right after the command action is executed
const afterExecute: AfterExecute<UserInput, ContractInput> = (context, input, deps) => async (result) => {
  deps.logger.info(
    `Contract deployed with address: ${result.responses[0].tx.address} at tx hash: ${result.responses[0].tx.hash}`,
  )
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.EXAMPLE,
    function: 'deploy',
    examples: [`${CATEGORIES.EXAMPLE}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
