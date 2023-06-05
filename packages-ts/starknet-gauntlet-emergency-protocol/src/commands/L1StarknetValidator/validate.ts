import { EVMExecuteCommandConfig, makeEVMExecuteCommand } from '@chainlink/evm-gauntlet'
import { CONTRACT_LIST, starknetValidatorContractLoader } from '../../lib/contracts'
import { CATEGORIES } from '../../lib/categories'

export interface UserInput {
  previousRoundId: number
  previousAnswer: number
  currentRoundId: number
  currentAnswer: number
}

type ContractInput = [
  previousRoundId: number,
  previousAnswer: number,
  currentRoundId: number,
  currentAnswer: number,
]

const makeContractInput = async ({
  previousRoundId,
  previousAnswer,
  currentRoundId,
  currentAnswer,
}: UserInput): Promise<ContractInput> => {
  return [previousRoundId, previousAnswer, currentRoundId, currentAnswer]
}

const validatePreviousRoundId = async (input) => {
  if (isNaN(Number(input.previousRoundId))) {
    throw new Error(`Invalid previousRoundId: ${input.previousRoundId}`)
  }
  return true
}

const validatePreviousAnswer = async (input) => {
  if (isNaN(Number(input.previousAnswer))) {
    throw new Error(`Invalid previousAnswer: ${input.previousAnswer}`)
  }
  return true
}

const validateCurrentRoundId = async (input) => {
  if (isNaN(Number(input.currentRoundId))) {
    throw new Error(`Invalid currentRoundId: ${input.currentRoundId}`)
  }
  return true
}

const validateCurrentAnswer = async (input) => {
  if (isNaN(Number(input.currentAnswer))) {
    throw new Error(`Invalid currentAnswer: ${input.currentAnswer}`)
  }
  return true
}

const makeUserInput = async (flags): Promise<UserInput> => {
  if (flags.input) return flags.input as UserInput
  const { previousRoundId, previousAnswer, currentRoundId, currentAnswer } = flags
  return {
    previousRoundId,
    previousAnswer,
    currentRoundId,
    currentAnswer,
  }
}

const commandConfig: EVMExecuteCommandConfig<UserInput, ContractInput> = {
  contractId: CONTRACT_LIST.STARKNET_VALIDATOR,
  category: CATEGORIES.STARKNET_VALIDATOR,
  action: 'validate',
  internalFunction: 'validate',
  ux: {
    description:
      'Validate the status by sending xDomain L2 tx to Starknet UptimeFeed. Caller must have access to validate.',
    examples: [
      `${CATEGORIES.STARKNET_VALIDATOR}:validate --previousRoundId=0 --previousAnswer=0 --currentRoundId=1 --currentAnswer=1 --network=<NETWORK> <YOUR_CONTRACT_ADDRESS>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [
    validateCurrentAnswer,
    validateCurrentRoundId,
    validatePreviousAnswer,
    validatePreviousRoundId,
  ],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
