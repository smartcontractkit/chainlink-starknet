import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

export interface UserInput {
  user: string
}

type ContractInput = [_user: string]

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.user]
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    user: flags.user,
  }
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'addAccess',
  ux: {
    description: 'Allow addres to access StarknetValidator',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:addAccess --user=0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4 0x6B5b7121C4F4B186e8C018a65CF379260B0Dba04 --network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
