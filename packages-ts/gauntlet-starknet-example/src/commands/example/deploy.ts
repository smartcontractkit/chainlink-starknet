import { ExecuteCommandConfig, makeExecuteCommand, BeforeExecute, AfterExecute } from '@chainlink/starknet-gauntlet'
import { DeployExampleBaseConfig, DeployExampleInput } from '@chainlink/gauntlet-contracts-example'
import { tokenContractLoader } from '../../lib/contracts'

type ContractInput = {}

const makeContractInput = async (input: DeployExampleInput): Promise<ContractInput> => {
  return {}
}

// This is a custom beforeExecute hook executed right before the command action is executed
const beforeExecute: BeforeExecute<DeployExampleInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info('About to deploy a Sample Contract')
}
// This is a custom afterExecute hook executed right after the command action is executed
const afterExecute: AfterExecute<DeployExampleInput, ContractInput> = (context, input, deps) => async (result) => {
  deps.logger.info(
    `Contract deployed with address: ${result.responses[0].tx.address} at tx hash: ${result.responses[0].tx.hash}`,
  )
}

const commandConfig: ExecuteCommandConfig<DeployExampleInput, ContractInput> = {
  ...DeployExampleBaseConfig,
  makeContractInput: makeContractInput,
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
    afterExecute,
  },
}

export default makeExecuteCommand(commandConfig)
