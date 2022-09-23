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
import { contractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  account: string
  amount: string
}

type ContractInput = [account: string, amount: Uint256]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    account: flags.account,
    amount: flags.amount,
  }
}

const validateAccount = async (input) => {
  if (!isValidAddress(input.account)) throw new Error(`Invalid account address: ${input.account}`)
  return true
}

const validateAmount = async (input) => {
  if (isNaN(Number(input.amount))) throw new Error(`Invalid amount: ${input.amount}`)
  return true
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [input.account, bnToUint256(input.amount)]
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
      `${CATEGORIES.TOKEN}:mint --network=<NETWORK> --account=<ACCOUNT> --amount=<AMOUNT> <CONTRACT_ADDRESS>`,
    ],
  },
  internalFunction: 'permissionedMint',
  makeUserInput,
  makeContractInput,
  validations: [validateAccount, validateAmount],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
