import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader } from '../../lib/contracts'
import { shortString } from 'starknet'

type UserInput = {
  owner: string
  billingAccessController: string
  linkToken: string
  decimals: number
  description: string
  maxAnswer: number
  minAnswer: number
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
  return {
    owner: flags.owner || env.account,
    maxAnswer: flags.maxSubmissionValue,
    minAnswer: flags.minSubmissionValue,
    decimals: flags.decimals,
    description: flags.name,
    billingAccessController: flags.billingAccessController || env.billingAccessController,
    linkToken: flags.link || env.link,
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return [
    input.owner,
    input.linkToken,
    new BN(input.minAnswer).toNumber(),
    new BN(input.maxAnswer).toNumber(),
    input.billingAccessController,
    new BN(input.decimals).toNumber(),
    shortString.encodeShortString(input.description),
  ]
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'deploy',
    examples: [
      `${CATEGORIES.OCR2}:deploy --network=<NETWORK> --billingAccessController=<ACCESS_CONTROLLER_CONTRACT> --link=<TOKEN_CONTRACT> --minSubmissionValue=1 --maxSubmissionValue=2 --decimals=3 --name="some feed name"`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
