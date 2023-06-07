import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
} from '@chainlink/starknet-gauntlet'
import { CATEGORIES } from '../../../lib/categories'
import { uptimeFeedContractLoader } from '../../../lib/contracts'

type Round = {
  round_id: string
  answer: string
  block_num: string
  started_at: string
  updated_at: string
}

type QueryResult = {
  latest_round_data: Round
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  let [latest_round_data] = results

  for (var key in latest_round_data) {
    if (latest_round_data.hasOwnProperty(key)) {
      latest_round_data[key] = BigInt(latest_round_data[key]).toString()
    }
  }

  return {
    toCompare: null,
    result: {
      latest_round_data: latest_round_data,
    },
  }
}

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.SEQUENCER_UPTIME_FEED,
    function: 'inspect',
    examples: [`${CATEGORIES.SEQUENCER_UPTIME_FEED}:inspect --network=<NETWORK>`],
  },
  queries: ['latest_round_data'],
  makeComparisionData,
  loadContract: uptimeFeedContractLoader,
}

export default makeInspectionCommand(commandConfig)
