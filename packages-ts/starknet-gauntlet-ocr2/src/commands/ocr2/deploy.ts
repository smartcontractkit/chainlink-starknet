import { CONTRACT_TYPES, ExecuteCommandConfig, makeExecuteCommand, ExecutionContext, getRDD } from '@chainlink/starknet-gauntlet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
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

  if (flags.rdd) {
    const rdd = getRDD(flags.rdd)
    const contractAddr = args[0]
    const aggregator = rdd[CONTRACT_TYPES.AGGREGATOR][contractAddr]
    return {
      maxAnswer: aggregator.maxSubmissionValue,
      minAnswer: aggregator.minSubmissionValue,
      decimals: aggregator.decimals,
      description: aggregator.name,
      billingAccessController: aggregator.billingAccessController || flags.billingAccessController || env.billingAccessController || '',
      linkToken: aggregator.linkToken || flags.linkToken || env.linkToken || '',
      owner: aggregator.owner || env.account,
    }
  }

  return {
    ...DeployOCR2.makeUserInput(flags, args, env),
    owner: flags.owner || env.account,
  } as UserInput
}

const makeContractInput = async (input: UserInput, ctx: ExecutionContext): Promise<ContractInput> => {
  return [
    input.owner,
    input.linkToken,
    input.minAnswer,
    input.maxAnswer,
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
