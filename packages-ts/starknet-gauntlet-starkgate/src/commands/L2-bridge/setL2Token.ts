import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { isValidAddress } from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { l2BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

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
  context: ExecutionContext,
): Promise<ContractInput> => {
  return [input.address]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!isValidAddress(input.address)) {
    throw new Error(`Invalid L2 token address: ${input.address}`)
  }

  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to set L2 Token of an L2 Bridge Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L2_BRIDGE,
  category: CATEGORIES.L2_BRIDGE,
  action: 'set_l2_token',
  ux: {
    description: 'Sets L1 token on an L2 token bridge',
    examples: [
      `${CATEGORIES.L2_BRIDGE}:set_l2_token --network=<NETWORK> --address=[L2_TOKEN_ADDRESS] [L2_BRIDGE_ADDRESS]`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: l2BridgeContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
