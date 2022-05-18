import { ExecuteCommandConfig, makeExecuteCommand, Validation } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '@chainlink/gauntlet-core'
import { tokenContractLoader } from '../../lib/contracts'
import { IncreaseBalance, IncreaseBalanceInput } from '@chainlink/gauntlet-contracts-example'

type ContractInput = [number]

const makeUserInput = async (flags, args): Promise<IncreaseBalanceInput> => {
  if (flags.input) return flags.input as IncreaseBalanceInput
  return {
    balance: flags.balance,
  }
}

const makeContractInput = async (input: IncreaseBalanceInput): Promise<ContractInput> => {
  return [Number(input.balance)]
}

const commandConfig: ExecuteCommandConfig<IncreaseBalanceInput, ContractInput> = {
  ...IncreaseBalance,
  makeContractInput: makeContractInput,
  loadContract: tokenContractLoader,
}

export default makeExecuteCommand(commandConfig)
