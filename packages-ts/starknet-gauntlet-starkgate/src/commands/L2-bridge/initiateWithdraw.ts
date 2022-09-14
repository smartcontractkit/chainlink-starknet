import { BeforeExecute, ExecuteCommandConfig, ExecutionContext, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { Uint256, bnToUint256 } from 'starknet/dist/utils/uint256'
import { CATEGORIES } from '../../lib/categories'
import { l2BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

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

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  return [input.recipient, bnToUint256(input.amount)]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!input.recipient) {
    throw new Error('Must specify --recipient of L1 Recipient')
  }

  if (!input.amount) {
    throw new Error('Must specify --amount to withdraw')
  }

  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to initiate withdraw to L1 with the following details:
    ${input.contract}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.L2_BRIDGE,
  category: CATEGORIES.L2_BRIDGE,
  action: 'initiate_withdraw',
  ux: {
    description: 'Initiates withdraw to L1',
    examples: [
      `${CATEGORIES.L2_BRIDGE}:initiate_withdraw --network=<NETWORK> --recipient=[L1_RECIPIENT_ADDRESS] --amount=[AMOUNT_TO_WITHDRAW] [L2_BRIDGE_ADDRESS]`,
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
