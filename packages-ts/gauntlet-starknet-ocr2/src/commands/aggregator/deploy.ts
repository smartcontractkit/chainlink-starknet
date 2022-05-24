import { ExecuteCommandConfig, makeExecuteCommand } from '@chainlink/gauntlet-starknet'
import { BN } from '@chainlink/gauntlet-core/dist/utils'
import { CATEGORIES } from '../../lib/categories'
import { ocr2ContractLoader } from '../../lib/contracts'
import { number } from 'starknet'

type UserInput = {
  billingAccessController: string
  linkToken: string
  decimals: number
  description: string
  maxAnswer: number
  minAnswer: number
}

type ContractInput = {
  // owner: string
  // link: string
  // min_answer: string
  // max_answer: string
  // billing_access_controller: string
  // decimals: string
  // description: string
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
  return [
    process.env.PUBLIC_KEY,
    process.env.LINK,
    new BN(input.minAnswer),
    new BN(input.maxAnswer),
    input.billingAccessController,
    new BN(input.decimals),
    input.description,
  ]
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
