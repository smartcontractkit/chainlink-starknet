import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/starknet-gauntlet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { ocr2ContractLoader } from '../../lib/contracts'
import { toFelt } from '../../lib/encoding'
import { shortString, number } from 'starknet'
import { DeployOCR2, DeployOCR2Input } from '@chainlink/gauntlet-contracts-ocr2'

export interface UserInput extends DeployOCR2Input {
  owner: string
}

type ContractInput = [
  owner: string,
  link: string,
  min_answer: BN,
  max_answer: BN,
  billing_access_controller: string,
  decimals: number,
  description: string,
]

const makeUserInput = async (flags, args, env): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    ...DeployOCR2.makeUserInput(flags, args, env),
    owner: flags.owner || env.account,
  } as UserInput
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [
    input.owner,
    input.linkToken,
    toFelt(input.minAnswer),
    toFelt(input.maxAnswer),
    input.billingAccessController,
    new BN(input.decimals).toNumber(),
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
