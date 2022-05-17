import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '@chainlink/gauntlet-core'
import { tokenContractLoader } from '../../lib/contracts'

type UserInput = {
  address: string
}

type ContractInput = {}

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return {}
}

const validate: Validation<UserInput> = async (input) => {
  return true
}

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
  contractId: 'exmaple',
  category: CATEGORIES.EXAMPLE,
  action: 'deploy',
  ux: {
    description: 'A sample contract deployment command',
    examples: [`${CATEGORIES.EXAMPLE}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
