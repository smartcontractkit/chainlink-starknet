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
    description: 'Transfers ownership of the StarknetValidator contract',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:transfer_ownership --to=0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4 0xAD6F411BF8559002CC9800A2E9aA87A0ff1b464e --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
