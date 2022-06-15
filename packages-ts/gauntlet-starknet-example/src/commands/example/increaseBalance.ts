import { ExecuteCommandConfig, makeExecuteCommand, Validation } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader } from '../../lib/contracts'

type UserInput = {
  balance: number
}

type ContractInput = [number]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    balance: flags.balance,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [Number(input.balance)]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.EXAMPLE,
    function: 'increase_balance',
    examples: ['token:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>'],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: tokenContractLoader,
}

export default makeExecuteCommand(commandConfig)
