import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
  InspectUserInput,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../lib/categories'
import { aggregatorConsumerLoader } from '../../lib/contracts'
import { uint256 } from 'starknet'

type UserInput = {}

type ContractInput = null

type QueryResult = {
  roundId: string
  answer: string
  decimals: string
}

const makeUserInput = async (flags, args): Promise<InspectUserInput<UserInput, null>> => {
  if (flags.input) return flags.input as InspectUserInput<UserInput, null>

  return {
    input: {},
  }
}

const makeContractInput = async (input: UserInput): Promise<ContractInput[]> => {
  return []
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const [round, decimals] = results
  const { round_id, answer } = round[0]

  // workaround proxy returning incorrect round id (0x1 in high register uint256)
  // todo: remove/adjust when figured out
  const roundIdUint256 = uint256.bnToUint256(round_id)
  roundIdUint256.high = 0
  const roundId = uint256.uint256ToBN(roundIdUint256)

  return {
    toCompare: null,
    result: {
      roundId: roundId.toString(),
      answer: answer.toString(),
      decimals: decimals[0].toString(),
    },
  }
}

const commandConfig: InspectCommandConfig<UserInput, ContractInput, null, QueryResult> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'inspect:consumer',
    examples: [
      `${CATEGORIES.OCR2}:inspect:consumer --network=<NETWORK> <AGGREGATOR_CONSUMER_ADDRESS>`,
    ],
  },
  queries: ['readLatestRound', 'readDecimals'],
  makeUserInput,
  makeContractInput,
  makeComparisionData,
  loadContract: aggregatorConsumerLoader,
}

export default makeInspectionCommand(commandConfig)
