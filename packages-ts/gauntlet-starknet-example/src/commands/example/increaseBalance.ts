import { ExecuteCommandConfig, makeExecuteCommand, Validation } from '@chainlink/gauntlet-starknet'
import { CATEGORIES } from '@chainlink/gauntlet-core'
import { tokenContractLoader } from '../../lib/contracts'

type UserInput = {
  balance: number
}

type ContractInput = [number]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    balance: flags.balance,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [Number(input.balance)]
}

const validate: Validation<UserInput> = async (input) => {
  return true
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: 'exmaple',
  category: CATEGORIES.EXAMPLE,
  action: 'increase_balance',
  ux: {
    description: 'A simple exmaple contract call - in this case increase_balance',
    examples: ['token:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>'],
  },
  makeUserInput,
  makeContractInput,
  validations: [validate],
  loadContract: tokenContractLoader,
}

export default makeExecuteCommand(commandConfig)
