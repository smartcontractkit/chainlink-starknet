import { ExecuteCommandConfig, makeExecuteCommand, Validation } from '@chainlink/gauntlet-starknet'
import { ContractInput } from '@chainlink/gauntlet-core'
import { TransferLink, TransferLinkInput } from '@chainlink/gauntlet-contracts-link'
import { tokenContractLoader } from '../../lib/contracts'

const commandConfig: ExecuteCommandConfig<TransferLinkInput, ContractInput> = {
  ...TransferLink,
  loadContract: tokenContractLoader,
}

export default makeExecuteCommand(commandConfig)
