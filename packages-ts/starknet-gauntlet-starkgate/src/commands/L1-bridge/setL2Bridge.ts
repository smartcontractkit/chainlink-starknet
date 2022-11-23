import {
  EVMExecuteCommandConfig,
  EVMExecutionContext,
  makeEVMExecuteCommand,
} from '@chainlink/evm-gauntlet'
import { isValidAddress } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { l1BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  address: string
}

type ContractInput = [address: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    address: flags.address,
  }
}

const makeContractInput = async (
  input: UserInput,
  context: EVMExecutionContext,
): Promise<ContractInput> => {
  return [input.address]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!isValidAddress(input.address)) {
    throw new Error(`Invalid address of L2 bridge: ${input.address}`)
  }
  return true
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_BRIDGE,
  category: CATEGORIES.L1_BRIDGE,
  action: 'set_l2_bridge',
  internalFunction: 'setL2TokenBridge',
  ux: {
    description: 'Sets L2 bridge',
    examples: [
      `${CATEGORIES.L1_BRIDGE}:set_l2_bridge --network=<NETWORK> --address=[L2_BRIDGE_ADDRESS] [L1_BRIDGE_PROXY_ADDRESS]`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: l1BridgeContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
