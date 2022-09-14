import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { number } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'
import { validateThreshold as validateThresholdWithSigners } from './deploy'

type UserInput = {
  threshold: string
}

type ContractInput = [threshold: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    threshold: flags.threshold,
  }
}

const validateThreshold = async (input: UserInput, context: ExecutionContext) => {
  const signers = (await context.contract.get_signers()).signers
  return validateThresholdWithSigners({ ...input, signers })
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [number.toFelt(input.threshold)]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger
    .info(`About to set a new threshold on the Multisig Contract (${context.contractAddress}) with the following details:
    - New Threshold: ${input.user.threshold}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    description: 'Set threshold of the multisig account',
    examples: [
      `${CATEGORIES.MULTISIG}:set_threshold --network=<NETWORK> --threshold=<MIN_APPROVALS> <MULTISIG_ADDRESS>`,
    ],
  },
  category: CATEGORIES.MULTISIG,
  contractId: CATEGORIES.MULTISIG,
  action: 'set_threshold',
  internalFunction: 'set_threshold',
  makeUserInput,
  makeContractInput,
  validations: [validateThreshold],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
