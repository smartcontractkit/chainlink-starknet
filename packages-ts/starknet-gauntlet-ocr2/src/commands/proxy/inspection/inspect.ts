import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
} from '@chainlink/starknet-gauntlet'
import { validateAndParseAddress, shortString } from 'starknet'
import { CATEGORIES } from '../../../lib/categories'
import { ocr2ProxyLoader } from '../../../lib/contracts'

type QueryResult = {
  round: any
  aggregator: string
  phaseId: string
  description: string
  decimals: number
  typeAndVersion: string
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  return {
    toCompare: null,
    result: {
      round: results[0][0],
      aggregator: validateAndParseAddress(results[1][0]),
      phaseId: results[2][0].toString(10),
      description: shortString.decodeShortString(results[3][0].toJSON()),
      decimals: results[4][0].toNumber(),
      typeAndVersion: shortString.decodeShortString(results[5][0].toJSON()),
    },
  }
}

// TODO: make inspection for proposed aggregator
// causes sequencer to throw 500 error and whole inspection to fail if proposed aggregator not set

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.PROXY,
    function: 'inspect',
    examples: ['yarn gauntlet proxy:inspect --network=<NETWORK> <PROXY>'],
  },
  queries: [
    'latest_round_data',
    'aggregator',
    'phase_id',
    'description',
    'decimals',
    'type_and_version',
  ],
  makeComparisionData,
  loadContract: ocr2ProxyLoader,
}

export default makeInspectionCommand(commandConfig)
