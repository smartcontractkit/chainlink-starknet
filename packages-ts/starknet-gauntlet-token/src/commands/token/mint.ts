import {
  BeforeExecute,
  ExecuteCommandConfig,
  makeExecuteCommand,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { uint256 } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { tokenContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  recipient: string
  amount: string
}

type ContractInput = [recipient: string, amount: any]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    recipient: flags.recipient,
    amount: flags.amount,
  }
}

const validateRecipient = async (input) => {
  if (!isValidAddress(input.recipient))
    throw new Error(`Invalid recipient address: ${input.recipient}`)
  return true
}

const validateAmount = async (input) => {
  if (isNaN(Number(input.amount))) throw new Error(`Invalid amount: ${input.amount}`)
  return true
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.recipient, uint256.bnToUint256(input.amount)]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to mint an ERC20 Token Contract with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.TOKEN,
  category: CATEGORIES.TOKEN,
  action: 'mint',
  ux: {
    description: 'Mints a set amount of tokens from contract to recipient',
    examples: [
      `${CATEGORIES.TOKEN}:mint --network=<NETWORK> --recipient=<ACCOUNT> --amount=<AMOUNT> <CONTRACT_ADDRESS>`,
    ],
  },
  internalFunction: 'permissionedMint',
  makeUserInput,
  makeContractInput,
  validations: [validateRecipient, validateAmount],
  loadContract: tokenContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
