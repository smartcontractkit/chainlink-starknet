import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  makeExecuteCommand,
  Validation,
  isValidAddress,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { accountContractLoader, CONTRACT_LIST } from '../../lib/contracts'

type UserInput = {
  classHash?: string
}

type ContractInput = []

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    classHash: flags.classHash,
  }
}

const validateClassHash = async (input) => {
  if (isValidAddress(input.classHash) || input.classHash === undefined) {
    return true
  }
  throw new Error(`Invalid Class Hash: ${input.classHash}`)
}

const makeContractInput = async (
  input: UserInput,
  context: ExecutionContext,
): Promise<ContractInput> => {
  return []
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to deploy an Argent Account Contract`)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.ACCOUNT,
  category: CATEGORIES.ACCOUNT,
  action: 'deploy',
  ux: {
    description: 'Deploys an Argent Labs Account contract',
    examples: [`${CATEGORIES.ACCOUNT}:deploy --classHash=<CLASS_HASH> --network=<NETWORK>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateClassHash],
  loadContract: accountContractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
