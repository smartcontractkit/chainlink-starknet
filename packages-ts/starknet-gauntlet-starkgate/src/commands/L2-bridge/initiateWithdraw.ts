import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { isValidAddress } from '@chainlink/evm-gauntlet'
import { uint256 } from 'starknet'
import { utils } from 'ethers'
import { CATEGORIES } from '../../lib/categories'
import { l2BridgeContractLoader, CONTRACT_LIST } from '../../lib/contracts'

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

const makeContractInput = async (
  input: UserInput,
  context: ExecutionContext,
): Promise<ContractInput> => {
  const amount = utils.parseEther(input.amount).toString()
  return [input.recipient, uint256.bnToUint256(amount)]
}

const validateInput = async (input: UserInput): Promise<boolean> => {
  if (!isValidAddress(input.recipient)) {
    throw new Error(`Invalid L1 recipient: ${input.recipient}`)
  }

  if (isNaN(Number(input.amount))) {
    throw new Error(`Invalid amount: ${input.amount}`)
  }

  return true
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
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
      `${CATEGORIES.L2_BRIDGE}:initiate_withdraw --network=<NETWORK> --recipient=[L1_RECIPIENT_ADDRESS] --amount=[AMOUNT_IN_LINK] [L2_BRIDGE_ADDRESS]`,
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
