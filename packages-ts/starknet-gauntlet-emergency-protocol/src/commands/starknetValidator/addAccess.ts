import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'
import { isValidAddress } from '@chainlink/starknet-gauntlet'

export interface UserInput {
  address: string
}

type ContractInput = [_user: string]

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.address]
}

const validateAddress = async (input) => {
  if (!isValidAddress(input.address)) {
    throw new Error(`Invalid Address: ${input.address}`)
  }
  return true
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    address: flags.address,
  }
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'add_access',
  internalFunction: 'addAccess',
  ux: {
    description: 'Allow address to access StarknetValidator',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:add_access --address=<ADDRESS> --network=<NETWORK> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateAddress],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
