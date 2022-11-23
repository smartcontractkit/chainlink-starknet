import {
  EVMExecuteCommandConfig,
  EVMExecutionContext,
  makeEVMExecuteCommand,
} from '@chainlink/evm-gauntlet'
import { BigNumber, utils } from 'ethers'
import { CATEGORIES } from '../../lib/categories'
import { l1BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  amount: string
}

type ContractInput = [amount: BigNumber]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    amount: flags.amount,
  }
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (isNaN(Number(input.amount))) {
    throw new Error(`Invalid amount: ${input.amount}`)
  }

  return true
}

const makeContractInput = async (
  input: UserInput,
  context: EVMExecutionContext,
): Promise<ContractInput> => {
  return [utils.parseEther(input.amount)]
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_BRIDGE,
  category: CATEGORIES.L1_BRIDGE,
  action: 'set_max_total_balance',
  internalFunction: 'setMaxTotalBalance',
  ux: {
    description: 'Sets Max Total Balance for the L1 Bridge',
    examples: [
      `${CATEGORIES.L1_BRIDGE}:set_max_total_balance --network=<NETWORK> --amount=[AMOUNT_IN_LINK] [L1_BRIDGE_PROXY_ADDRESS]`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: l1BridgeContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
