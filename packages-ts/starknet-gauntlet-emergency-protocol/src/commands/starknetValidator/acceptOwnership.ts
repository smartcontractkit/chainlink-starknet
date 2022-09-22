import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

export type ContractInput = []

export interface UserInput {}

const makeContractInput = async (input: ContractInput): Promise<ContractInput> => {
  return input
}

const makeUserInput = async (flags): Promise<UserInput> => {
  return {}
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'accept_ownership',
  internalFunction: 'acceptOwnership',
  ux: {
    description: 'Accepts ownership of the StarknetValidator contract',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:accept_ownership 0xAD6F411BF8559002CC9800A2E9aA87A0ff1b464e --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
