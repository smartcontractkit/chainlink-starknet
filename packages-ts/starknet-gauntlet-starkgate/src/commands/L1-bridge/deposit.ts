import {
  EVMExecuteCommandConfig,
  EVMExecutionContext,
  makeEVMExecuteCommand,
} from '@chainlink/evm-gauntlet'
import { isValidAddress } from '@chainlink/starknet-gauntlet'
import { BigNumber, utils } from 'ethers'
import { CATEGORIES } from '../../lib/categories'
import { l1BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  amount: string
  recipient: string
}

type ContractInput = [amount: BigNumber, recipient: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    amount: flags.amount,
    recipient: flags.recipient,
  }
}

const makeContractInput = async (
  input: UserInput,
  context: EVMExecutionContext,
): Promise<ContractInput> => {
  return [utils.parseEther(input.amount), input.recipient]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (isNaN(Number(input.amount))) {
    throw new Error(`Invalid amount: ${input.amount}`)
  }

  if (!isValidAddress(input.recipient)) {
    throw new Error(`Invalid address of L2 recipient: ${input.recipient}`)
  }

  return true
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L1_BRIDGE,
  category: CATEGORIES.L1_BRIDGE,
  action: 'deposit',
  ux: {
    description: 'Deposits funds to L1 bridge for L2 recipient',
    examples: [
      `${CATEGORIES.L1_BRIDGE}:deposit --network=<NETWORK> --recipient=[L2_RECIPIENT_ADDRESS] --amount=[AMOUNT_IN_LINK] [L1_BRIDGE_PROXY_ADDRESS]`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateInput],
  loadContract: l1BridgeContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
