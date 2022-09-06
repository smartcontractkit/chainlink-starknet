import {
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  Validation,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { Uint256 } from 'starknet/dist/utils/uint256'
import { bnToUint256 } from 'starknet/dist/utils/uint256'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  recipient: string
  amount: string
}

type ContractInput = [recipient: string, amount: Uint256]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    recipient: flags.recipient,
    amount: flags.amount,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.recipient, bnToUint256(input.amount)]
}

const validateRecipient = async (input) => {
  if (!isValidAddress(input.recipient)) throw new Error(`Invalid recipient address: ${input.recipient}`)
  return true
}

const validateAmount = async (input) => {
  if (isNaN(Number(input.amount))) throw new Error(`Invalid amount: ${input.amount}`)
  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to tranfer ${input.user.amount} ERC20 tokens to ${input.user.recipient}`)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.TOKEN,
  category: CATEGORIES.TOKEN,
  action: 'transfer',
  ux: {
    description: 'Transfers a set amount of tokens from caller to recipient',
    examples: [
      `${CATEGORIES.TOKEN}:transfer --network=<NETWORK> --recipient=<RECIPIENT> --amount=<AMOUNT> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateRecipient, validateAmount],
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
