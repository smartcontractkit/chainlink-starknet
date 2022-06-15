import { BeforeExecute, ExecuteCommandConfig, ExecutionContext, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { number } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'
import { validateThreshold as validateThresholdWithOwners } from './deploy'

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
  const owners = (await context.contract.get_owners()).owners
  return validateThresholdWithOwners({ ...input, owners })
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [number.toFelt(input.threshold)]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger
    .info(`About to set a new threshold on the Multisig Contract (${context.contractAddress}) with the following details:
    - New Threshold: ${input.user.threshold}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.MULTISIG,
    function: 'set_threshold',
    examples: [
      `${CATEGORIES.MULTISIG}:set_threshold --network=<NETWORK> --threshold=<MIN_APPROVALS> <MULTISIG_ADDRESS>`,
    ],
  },
  internalFunction: 'set_confirmations_required',
  makeUserInput,
  makeContractInput,
  validations: [validateThreshold],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
