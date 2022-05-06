import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { Validation } from '@chainlink/gauntlet-starknet/dist/commands/command'
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
  console.log('validating token example input')
  return true
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
}

export default makeExecuteCommand(commandConfig)
