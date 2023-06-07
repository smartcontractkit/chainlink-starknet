import {
  CONTRACT_TYPES,
  ExecuteCommandConfig,
  makeExecuteCommand,
  ExecutionContext,
  getRDD,
} from '@chainlink/starknet-gauntlet'
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
      billingAccessController: flags.billingAccessController || env.BILLING_ACCESS_CONTROLLER || '',
      linkToken: flags.linkToken || env.LINK || '',
      owner: flags.owner || env.ACCOUNT,
    }
  }

  flags.min_answer = parseInt(flags.min_answer)
  flags.max_answer = parseInt(flags.max_answer)
  flags.decimals = parseInt(flags.decimals)

  return {
    ...DeployOCR2.makeUserInput(flags, args, env),
    owner: flags.owner || env.account,
  } as UserInput
}

const makeContractInput = async (
  input: UserInput,
  ctx: ExecutionContext,
): Promise<ContractInput> => {
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
  ux: {
    description: 'Deploys OCR2 contract',
    examples: [
      `yarn gauntlet ocr2:deploy --network=<NETWORK> --billingAccessController=<ACCESS_CONTROLLER_CONTRACT> --minSubmissionValue=<MIN_VALUE> --maxSubmissionValue=<MAX_VALUE> --decimals=<DECIMALS> --name=<FEED_NAME> --link=<TOKEN_CONTRACT> --owner=<OWNER>`,
      `yarn gauntlet ocr2:deploy --network=<NETWORK> --rdd=<RDD_PATH> --billingAccessController=<ACCESS_CONTROLLER_CONTRACT> --link=<TOKEN_CONTRACT> --owner=<OWNER> <CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput: makeUserInput,
  makeContractInput: makeContractInput,
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
