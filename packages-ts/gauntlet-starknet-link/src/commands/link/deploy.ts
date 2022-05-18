import {
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  BeforeExecute,
  AfterExecute,
} from '@chainlink/gauntlet-starknet'
import { ContractInput } from '@chainlink/gauntlet-core'
import { tokenContractLoader } from '../../lib/contracts'
import { DeployLink, DeployLinkInput } from '@chainlink/gauntlet-contracts-link'

// This is a custom beforeExecute hook executed right before the command action is executed
const beforeExecute: BeforeExecute<DeployLinkInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info('About to deploy a Sample Contract')
  await deps.prompt('Continue?')
}
// This is a custom afterExecute hook executed right after the command action is executed
const afterExecute: AfterExecute<DeployLinkInput, ContractInput> = (context, input, deps) => async (result) => {
  deps.logger.info(
    `Contract deployed with address: ${result.responses[0].tx.address} at tx hash: ${result.responses[0].tx.hash}`,
  )
}

const commandConfig: ExecuteCommandConfig<DeployLinkInput, ContractInput> = {
  ...DeployLink,
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
