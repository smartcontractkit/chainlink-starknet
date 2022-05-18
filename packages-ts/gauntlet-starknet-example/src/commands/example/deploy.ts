import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '@chainlink/gauntlet-core'
import { tokenContractLoader } from '../../lib/contracts'
import { DeployExample, DeployExampleInput } from '@chainlink/gauntlet-contracts-example'

type ContractInput = {}

const makeUserInput = async (flags, args): Promise<DeployExampleInput> => {
  if (flags.input) return flags.input as DeployExampleInput
  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: DeployExampleInput): Promise<ContractInput> => {
  return {}
}

const validate: Validation<DeployExampleInput> = async (input) => {
  return true
}

// This is a custom beforeExecute hook executed right before the command action is executed
const beforeExecute: BeforeExecute<DeployExampleInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info('About to deploy a Sample Contract')
  await deps.prompt('Continue?')
}
// This is a custom afterExecute hook executed right after the command action is executed
const afterExecute: AfterExecute<DeployExampleInput, ContractInput> = (context, input, deps) => async (result) => {
  deps.logger.info(
    `Contract deployed with address: ${result.responses[0].tx.address} at tx hash: ${result.responses[0].tx.hash}`,
  )
}

const commandConfig: ExecuteCommandConfig<DeployExampleInput, ContractInput> = {
  ...DeployExample,
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
