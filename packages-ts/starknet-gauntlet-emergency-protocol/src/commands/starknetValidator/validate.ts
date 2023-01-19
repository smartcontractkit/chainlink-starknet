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
      `${CATEGORIES.STARKNET_VALIDATOR}:validate --previousRoundId=0x0 --previousAnswer=0x0 --currentRoundId=0x1 --currentAnswer=0x1 <YOUR_CONTRACT_ADDRESS>--network=<NETWORK>`,
    ],
  },
  makeUserInput,
  makeContractInput,
  validations: [],
  loadContract: starknetValidatorContractLoader,
}

export default makeEVMExecuteCommand(commandConfig)
