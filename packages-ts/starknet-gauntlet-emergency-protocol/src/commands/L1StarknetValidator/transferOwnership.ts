import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

export interface UserInput {
  to: string
}

type ContractInput = [to: string]

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.to]
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    to: flags.to,
  }
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'transfer_ownership',
  internalFunction: 'transferOwnership',
  ux: {
    description:
      'Transfers ownership of the StarknetValidator contract. Should be called by current owner',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:transfer_ownership --to=<NEW_PROPOSED_OWNER> <CONTRACT_ADDRESS> --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
