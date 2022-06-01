import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  isValidAddress,
  makeExecuteCommand,
  Validation,
} from '@chainlink/gauntlet-starknet'
import { number } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'

type UserInput = {
  owners: string[]
  threshold: string
}

type ContractInput = [ownersLen: string, ...owners: string[], threshold: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    owners: flags.owners,
    threshold: flags.threshold,
  }
}

const validateThreshold: Validation<UserInput> = async (input) => {
  const threshold = Number(input.threshold)
  if (isNaN(threshold)) throw new Error('Threshold is not a number')
  if (threshold > input.owners.length)
    throw new Error(`Threshold is higher than owners length: ${threshold} > ${input.owners.length}`)
  return true
}

const validateOwners: Validation<UserInput> = async (input) => {
  const areValid = input.owners.every(isValidAddress)
  if (!areValid) throw new Error('Owners are not valid accounts')
  if (new Set(input.owners).size !== input.owners.length) throw new Error('Owners are not unique')
  return true
}

const makeContractInput = async (input: UserInput, context: ExecutionContext): Promise<ContractInput> => {
  return [
    number.toFelt(input.owners.length),
    ...input.owners.map((addr) => number.toFelt(number.toBN(addr))),
    number.toFelt(input.threshold),
  ]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (context, input, deps) => async () => {
  deps.logger.info(`About to deploy an Multisig Contract with the following details:
    - Owners: ${input.user.owners}
    - Threshold: ${input.user.threshold}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.MULTISIG,
    function: 'deploy',
    examples: [`${CATEGORIES.MULTISIG}:deploy --network=<NETWORK> --threshold=<MIN_APPROVALS> --owners=[OWNERS_LIST]`],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateOwners, validateThreshold],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
