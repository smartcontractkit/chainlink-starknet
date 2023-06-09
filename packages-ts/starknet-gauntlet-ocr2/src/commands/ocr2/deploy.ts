import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { ocr2ContractLoader } from '../../lib/contracts'
import { shortString } from 'starknet'
import { DeployOCR2, DeployOCR2Input } from '@chainlink/gauntlet-contracts-ocr2'
import { validateClassHash } from '../../lib/utils'

export interface UserInput extends DeployOCR2Input {
  owner: string
  classHash?: string
}

type ContractInput = [
  owner: string,
  link: string,
  min_answer: number,
  max_answer: number,
  billing_access_controller: string,
  decimals: number,
  description: string,
]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput

  flags.min_answer = parseInt(flags.min_answer)
  flags.max_answer = parseInt(flags.max_answer)
  flags.decimals = parseInt(flags.decimals)

  const classHash = flags.classHash
  const input = {
    ...DeployOCR2.makeUserInput(flags, args, env),
    owner: flags.owner || env.account,
  } as UserInput
  // DeployOCR2.validations does not allow input keys to be "false-y" so we only add classHash key if it is !== undefined
  if (classHash !== undefined) {
    input['classHash'] = classHash
  }
  return input
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [
    input.owner,
    input.linkToken,
    input.minAnswer,
    input.maxAnswer,
    input.billingAccessController,
    input.decimals,
    shortString.encodeShortString(input.description),
  ]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ...DeployOCR2,
  validations: [...DeployOCR2.validations, validateClassHash],
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
