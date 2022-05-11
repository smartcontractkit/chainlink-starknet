import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContract } from '../../lib/contracts'

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
const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async (signer) => {
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
    category: CATEGORIES.TOKEN,
    function: 'deploy',
    examples: [`${CATEGORIES.TOKEN}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  contract: tokenContract,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
