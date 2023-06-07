import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { ocr2ContractLoader } from '../../lib/contracts'
import { shortString } from 'starknet'
import { DeployOCR2, DeployOCR2Input } from '@chainlink/gauntlet-contracts-ocr2'

export interface UserInput extends DeployOCR2Input {
  owner: string
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

  return {
    ...DeployOCR2.makeUserInput(flags, args, env),
    owner: flags.owner || env.account,
  } as UserInput
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
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
