import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

export interface ContractInput {}

const makeContractInput = async (input: ContractInput): Promise<ContractInput> => {
  return input
}

const makeUserInput = async (flags): Promise<ContractInput> => {
  return {}
}

const commandConfig: EVMExecuteCommandConfig<ContractInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'acceptOwnership',
  ux: {
    description: 'Accepts ownership of the StarknetValidator contract',
    examples: [`${CATEGORIES.STARKNET_VALIDATOR}:acceptOwnership --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
