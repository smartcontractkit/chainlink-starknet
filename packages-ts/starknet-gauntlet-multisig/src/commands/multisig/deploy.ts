import {
  BeforeExecute,
  ExecuteCommandConfig,
  ExecutionContext,
  isValidAddress,
  makeExecuteCommand,
} from '@chainlink/starknet-gauntlet'
import { number } from 'starknet'
import { CATEGORIES } from '../../lib/categories'
import { contractLoader } from '../../lib/contracts'

type UserInput = {
  signers: string[]
  threshold: string
}

type ContractInput = [signersLen: string, ...signers: string[], threshold: string]

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  return {
    signers: flags.signers,
    threshold: flags.threshold,
  }
}

export const validateThreshold = async (input: UserInput) => {
  const threshold = Number(input.threshold)
  if (isNaN(threshold)) throw new Error('Threshold is not a number')
  if (threshold > input.signers.length)
    throw new Error(
      `Threshold is higher than signers length: ${threshold} > ${input.signers.length}`,
    )
  return true
}

export const validateSigners = async (input) => {
  const areValid = input.signers.every(isValidAddress)
  if (!areValid) throw new Error('Signers are not valid accounts')
  if (new Set(input.signers).size !== input.signers.length)
    throw new Error('Signers are not unique')
  return true
}

const makeContractInput = async (
  input: UserInput,
  context: ExecutionContext,
): Promise<ContractInput> => {
  return [
    number.toFelt(input.signers.length),
    ...input.signers.map((addr) => number.toFelt(number.toBN(addr))),
    number.toFelt(input.threshold),
  ]
}

const beforeExecute: BeforeExecute<UserInput, ContractInput> = (
  context,
  input,
  deps,
) => async () => {
  deps.logger.info(`About to deploy an Multisig Contract with the following details:
    - Signers: ${input.user.signers}
    - Threshold: ${input.user.threshold}
  `)
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CATEGORIES.MULTISIG,
  category: CATEGORIES.MULTISIG,
  action: 'deploy',
  ux: {
    description: 'Deploys Multisig Wallet',
    examples: [
      `${CATEGORIES.MULTISIG}:deploy --network=<NETWORK> --threshold=<MIN_APPROVALS> --signers=[SIGNERS_LIST]`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [validateSigners, validateThreshold],
  loadContract: contractLoader,
  hooks: {
    beforeExecute,
  },
}

export default makeExecuteCommand(commandConfig)
