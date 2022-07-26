import { BN } from '@chainlink/gauntlet-core/dist/utils'
import {
  AfterExecute,
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  Validation,
} from '@chainlink/starknet-gauntlet'
import { shortString } from 'starknet'
import { isContext } from 'vm'
import { CATEGORIES } from '../../lib/categories'
import { bridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  address?: string
}

type ContractInput = [address: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    address: flags.address,
  }
}

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  return [input.address]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!input.address) {
    throw new Error('Must supply --address of L1 Bridge')
  }
  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to set L1 Bridge of an L2 Bridge Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.BRIDGE,
  category: CATEGORIES.BRIDGE,
  action: 'set_l1_bridge',
  ux: {
    description: 'Sets L1 bridge on an L2 token bridge',
    examples: [
      `${CATEGORIES.BRIDGE}:set_l1_bridge --network=<NETWORK> --address=[L1_BRIDGE_ADDRESS] [L2_BRIDGE_ADDRESS]`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: bridgeContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
