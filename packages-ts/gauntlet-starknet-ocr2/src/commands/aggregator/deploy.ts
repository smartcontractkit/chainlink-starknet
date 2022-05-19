import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader } from '../../lib/contracts'

type UserInput = {
  billingAccessController: string
  linkToken: string
  decimals: number
  description: string
  maxAnswer: string
  minAnswer: string
}

type ContractInput = {
  link: string
  min_answer: number
  max_answer: number
  billing_access_controller: string
  decimals: number
  description: string
}

const makeUserInput = async (flags, args): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  return {
    maxAnswer: flags.maxSubmissionValue,
    minAnswer: flags.minSubmissionValue,
    decimals: flags.decimals,
    description: flags.name,
    billingAccessController: process.env.BILLING_ACCESS_CONTROLLER || '',
    linkToken: process.env.LINK || '',
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput> => {
  return {
    min_answer: new BN(input.minAnswer).toNumber(),
    max_answer: new BN(input.maxAnswer).toNumber(),
    decimals: new BN(input.decimals).toNumber(),
    billing_access_controller: input.billingAccessController,
    description: input.description,
    link: input.linkToken,
  }
}

const commandConfig: ExecuteCommandConfig<UserInput, ContractInput> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'deploy',
    examples: [`${CATEGORIES.OCR2}:deploy --network=<NETWORK> --address=<ADDRESS> <CONTRACT_ADDRESS>`],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: ocr2ContractLoader,
}

export default makeExecuteCommand(commandConfig)
